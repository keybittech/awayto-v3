#################################
#             TARGETS           #
#################################

build: $(LOG_DIR) ${SIGNING_TOKEN_FILE} ${KC_PASS_FILE} ${KC_USER_CLIENT_SECRET_FILE} ${KC_API_CLIENT_SECRET_FILE} ${PG_PASS_FILE} ${PG_WORKER_PASS_FILE} ${REDIS_PASS_FILE} ${AI_KEY_FILE} $(CERT_LOC) $(CERT_KEY_LOC) $(JAVA_TARGET) $(LANDING_TARGET) $(TS_TARGET) $(PROTO_GEN_FILES) $(PROTO_GEN_MUTEX) $(PROTO_GEN_MUTEX_FILES) $(GO_HANDLERS_REGISTER) $(GO_TARGET)

# certs, secrets, demo and backup dirs are not cleaned
.PHONY: clean
clean:
	rm -rf $(LOG_DIR) $(TS_BUILD_DIR) $(GO_GEN_DIR) \
		$(LANDING_BUILD_DIR) $(JAVA_TARGET_DIR) $(PLAYWRIGHT_CACHE_DIR)
	rm -f $(GO_TARGET) $(BINARY_TEST) $(TS_API_YAML) $(TS_API_BUILD) $(GO_HANDLERS_REGISTER)

$(LOG_DIR):
ifeq ($(DEPLOYING),)
	mkdir -p $(LOG_DIR)/db
	setfacl -m g:1000:rwx $(LOG_DIR)/db
	setfacl -d -m g:1000:rwx $(LOG_DIR)/db
endif

$(CERT_LOC) $(CERT_KEY_LOC):
ifeq ($(DEPLOYING),)
	echo $(H_LOGIN) $(DEPLOYING) $(shell whoami) $(CERT_KEY_LOC) $(CERT_LOC)
	mkdir -p $(@D)
	chmod 750 $(@D)
	openssl req -nodes -new -x509 -keyout $(CERT_KEY_LOC) -out $(CERT_LOC) -days 365 -subj "/CN=${APP_HOST_NAME}"
	chmod 600 $(CERT_KEY_LOC) $(CERT_LOC)
	echo "pwd required to update cert chain"
	sudo cp $(CERT_LOC) /usr/local/share/ca-certificates
	sudo update-ca-certificates
endif

${SIGNING_TOKEN_FILE} ${KC_PASS_FILE} ${KC_USER_CLIENT_SECRET_FILE} ${KC_API_CLIENT_SECRET_FILE} ${PG_PASS_FILE} ${PG_WORKER_PASS_FILE} ${REDIS_PASS_FILE}:
	@mkdir -p $(@D)
	install -m 640 /dev/null $@
	openssl rand -hex 64 > $@ | tr -d '\n'
ifeq ($(DEPLOYING),true)
	@chown -R $(H_LOGIN):$(H_GROUP) $(@D)
else
	@chown -R $(shell whoami):1000 $(@D)
endif
	@chmod -R 750 $(@D)

${AI_KEY_FILE}:
	install -m 640 /dev/null $@
	echo "Provide an AiStudio API key if desired, or just press Enter"
	read -s AI_KEY; echo "$$AI_KEY" > $@

$(JAVA_TARGET): $(shell find $(JAVA_SRC)/{src,themes,pom.xml} -type f)
	rm -rf $(JAVA_SRC)/target
	mkdir $(@D)
	cp $(JAVA_SRC)/junixsocket-selftest-2.10.1-jar-with-dependencies.jar $(JAVA_TARGET_DIR)/
	mvn -f $(JAVA_SRC) install

$(LANDING_SRC)/config.yaml:
	sed -e 's&project-title&${PROJECT_TITLE}&g; s&last-updated&$(shell date +%Y-%m-%d)&g; s&app-host-url&${APP_HOST_URL}&g;' "$(LANDING_SRC)/config.yaml.template" > "$(LANDING_SRC)/config.yaml"

# using npm here as pnpm symlinks just hugo and doesn't build correctly 
$(LANDING_TARGET): $(LANDING_SRC)/config.yaml $(LANDING_SRC)/config.yaml.template $(shell find $(LANDING_SRC)/{assets,content,layouts,static,package-lock.json} -type f)
	npm --prefix ${LANDING_SRC} i
	npm run --prefix ${LANDING_SRC} build

$(TS_API_BUILD): $(shell find proto/ -type f)
	protoc --proto_path=proto \
		--experimental_allow_proto3_optional \
		--openapi_out=$(TS_SRC) \
		$(PROTO_FILES)
	npx -y @rtk-query/codegen-openapi $(TS_CONFIG_API)

$(TS_SRC)/.env.local: $(TS_SRC)/.env.template
	sed -e 's&project-title&${PROJECT_TITLE}&g; s&app-host-url&${APP_HOST_URL}&g; s&app-host-name&${APP_HOST_NAME}&g; s&turn-name&${TURN_NAME}&g; s&turn-pass&${TURN_PASS}&g; s&allowed-file-ext&${ALLOWED_FILE_EXT}&g; s&ai-enabled&$(AI_ENABLED)&g;' "$(TS_SRC)/.env.template" > "$(TS_SRC)/.env.local"

$(TS_TARGET): $(TS_SRC)/.env.local $(TS_API_BUILD) $(shell find $(TS_SRC)/{src,public,package.json,index.html,vite.config.ts} -type f) $(shell find proto/ -type f)
	pnpm --dir $(TS_SRC) i
	pnpm run --dir $(TS_SRC) build

$(PROTO_GEN_FILES): $(PROTO_FILES) 
	@mkdir -p $(@D) $(GO_GEN_DIR)
	rm $(GO_GEN_DIR)/*.pb.go || true
	protoc --proto_path=proto \
		--experimental_allow_proto3_optional \
		--go_out=$(GO_GEN_DIR) \
		--go_opt=module=${PROJECT_REPO}/$(GO_GEN_DIR) \
		$(PROTO_FILES)

$(PROTO_GEN_MUTEX): $(GO_PROTO_MUTEX_CMD_DIR)/main.go
	$(GO) build -C $(GO_PROTO_MUTEX_CMD_DIR) -o $@ .

$(PROTO_GEN_MUTEX_FILES): $(PROTO_GEN_MUTEX) $(PROTO_FILES)
	protoc --proto_path=proto \
		--plugin=protoc-gen-mutex=$(PROTO_GEN_MUTEX) \
		--mutex_out=$(GO_GEN_DIR) \
		--mutex_opt=module=${PROJECT_REPO}/$(GO_GEN_DIR) \
		$(PROTO_FILES)

$(GO_HANDLERS_REGISTER): $(GO_HANDLERS_REGISTER_CMD_DIR)/main.go $(PROTO_FILES)
	$(GO) run -C $(GO_HANDLERS_REGISTER_CMD_DIR) ./...

$(GO_TARGET): $(GO_FILES)
	$(call set_local_unix_sock_dir)
	$(GO) build -C $(GO_SRC) -o $(GO_TARGET) .
