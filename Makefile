SHELL := /bin/bash

GO_VERSION=go1.25.1.linux-amd64
NODE_VERSION=v22.13.1

# manually manage path for makefile use
export PATH := /home/$(shell whoami)/.nvm/versions/node/$(NODE_VERSION)/bin:/home/$(shell whoami)/go/bin:/usr/local/go/bin:${PATH}

ENVFILE?=./.env

$(shell if [ ! -f $(ENVFILE) ]; then install -m 600 .env.template $(ENVFILE); fi)

include $(ENVFILE)
export $(shell sed 's/=.*//' $(ENVFILE))

export UID := $(shell id -u)
export GID := $(shell id -g)

# go build output

BINARY_TEST=$(BINARY_NAME).test
BINARY_SERVICE=$(BINARY_NAME).service

#################################
#          SOURCE DIRS          #
#################################

JAVA_SRC=java
TS_SRC=ts
GO_SRC=go
LANDING_SRC=landing

#################################
#           LOCAL DIRS          #
#################################

JAVA_TARGET_DIR=java/target
JAVA_THEMES_DIR=java/themes
LANDING_BUILD_DIR=landing/public
TS_BUILD_DIR=ts/build
TS_PUBLIC_DIR=ts/public
LOG_DIR=log
HOST_LOCAL_DIR=sites/${PROJECT_PREFIX}
GO_GEN_DIR=$(GO_SRC)/pkg/types
GO_CMD_DIR=$(GO_SRC)/cmd
GO_VAULT_WASM_DIR=$(GO_CMD_DIR)/crypto/vault
GO_PROTO_MUTEX_CMD_DIR=$(GO_CMD_DIR)/generate/proto_mutex
GO_HANDLERS_REGISTER_CMD_DIR=$(GO_CMD_DIR)/generate/handlers_register
GO_API_DIR=$(GO_SRC)/pkg/api
GO_CLIENTS_DIR=$(GO_SRC)/pkg/clients
GO_CRYPTO_DIR=$(GO_SRC)/pkg/crypto
export GO_HANDLERS_DIR=$(GO_SRC)/pkg/handlers
GO_UTIL_DIR=$(GO_SRC)/pkg/util
GO_INTEGRATIONS_DIR=$(GO_SRC)/integrations
GO_PLAYWRIGHT_DIR=$(GO_SRC)/playwright
# GO_INTERFACES_DIR=$(GO_SRC)/pkg/interfaces
export PLAYWRIGHT_CACHE_DIR=working/playwright
DEMOS_DIR=demos/final

#################################
#          TARGET VARS          #
#################################

JAVA_TARGET=$(JAVA_TARGET_DIR)/kc-custom.jar
LANDING_TARGET=$(LANDING_BUILD_DIR)/index.html
TS_TARGET=$(TS_BUILD_DIR)/index.html

GO_FILE_DIRS=$(GO_SRC) $(GO_HANDLERS_REGISTER_CMD_DIR) $(GO_GEN_DIR) $(GO_API_DIR) $(GO_CLIENTS_DIR) $(GO_CRYPTO_DIR) $(GO_HANDLERS_DIR) $(GO_UTIL_DIR)
GO_FILES=$(foreach dir,$(GO_FILE_DIRS),$(wildcard $(dir)/*.go))
GO_TARGET=${PWD}/$(GO_SRC)/$(BINARY_NAME)
export TS_VAULT_WASM=$(TS_PUBLIC_DIR)/lib.wasm
export GO_HANDLERS_REGISTER=$(GO_HANDLERS_DIR)/register.go
# GO_INTERFACES_FILE=$(GO_INTERFACES_DIR)/interfaces.go
# GO_MOCK_TARGET=$(GO_INTERFACES_DIR)/mocks.go

PROTO_FILES:=$(wildcard proto/*.proto)
PROTO_GEN_FILES:=$(patsubst proto/%.proto,$(GO_GEN_DIR)/%.pb.go,$(PROTO_FILES))
PROTO_GEN_MUTEX=$(PROJECT_DIR)/working/protoc-gen-mutex
PROTO_GEN_MUTEX_FILES:=$(patsubst proto/%.proto,$(GO_GEN_DIR)/%_mutex.pb.go,$(PROTO_FILES))

TS_API_YAML=ts/openapi.yaml
TS_API_BUILD=ts/src/hooks/api.ts
TS_CONFIG_API=ts/openapi-config.json

#################################
#            BACKUPS            #
#################################

BACKUP_DIR=backups
DB_BACKUP_DIR=$(BACKUP_DIR)/db
CERT_BACKUP_DIR=$(BACKUP_DIR)/certs
DOCKER_REDIS_CID=$(shell docker ps -aqf "name=redis")
DOCKER_REDIS_EXEC=docker exec -i $(DOCKER_REDIS_CID) redis-cli --pass $$(cat $$REDIS_PASS_FILE)
DOCKER_DB_CID=$(shell docker ps -aqf "name=db")
DOCKER_DB_EXEC=docker exec --user postgres -it
DOCKER_DB_CMD=docker exec --user postgres -i

#################################
#          DEPLOY PROPS         #
#################################

DEPLOY_SCRIPTS=deploy/scripts
DB_SCRIPTS=$(DEPLOY_SCRIPTS)/db
DEV_SCRIPTS=$(DEPLOY_SCRIPTS)/dev
DOCKER_SCRIPTS=$(DEPLOY_SCRIPTS)/docker
DEPLOY_HOST_SCRIPTS=$(DEPLOY_SCRIPTS)/host
CRON_SCRIPTS=$(DEPLOY_SCRIPTS)/cron
DOCKER_COMPOSE_SCRIPT=$(DEPLOY_SCRIPTS)/docker/docker-compose.yml
AUTH_SCRIPTS=$(DEPLOY_SCRIPTS)/auth
AUTH_INSTALL_SCRIPT=$(AUTH_SCRIPTS)/install.sh

H_ETC_DIR=/etc/${PROJECT_PREFIX}

H_LOGIN=${HOST_OPERATOR}login
H_GROUP=${PROJECT_PREFIX}group
H_SIGN=$(H_LOGIN)@${APP_HOST}
SSH=tailscale ssh $(H_SIGN)
TSIPV4=$(shell tailscale ip -4 ${APP_HOST})

CURRENT_USER:=$(shell whoami)
DEPLOYING=$(if $(filter ${HOST_OPERATOR}login,$(CURRENT_USER)),true,)

define if_deploying
$(if $(DEPLOYING),$(1),$(2))
endef

DEPLOYING_LOCALLY=$(shell test -f "$(HOST_LOCAL_DIR)/.local" && echo true)

define if_deploying_locally
$(if $(DEPLOYING_LOCALLY),$(1),$(2))
endef

CURRENT_APP_HOST_NAME=$(call if_deploying,${DOMAIN_NAME},localhost:${GO_HTTPS_PORT})
CURRENT_CERTS_DIR=$(call if_deploying,/etc/letsencrypt/live/${DOMAIN_NAME},${PWD}/${CERTS_DIR})
CURRENT_CERT_LOC=$(CURRENT_CERTS_DIR)/cert.pem
CURRENT_CERT_KEY_LOC=$(CURRENT_CERTS_DIR)/privkey.pem
CURRENT_PROJECT_DIR=$(call if_deploying,/etc/${PROJECT_PREFIX},${PWD})
CURRENT_LOG_DIR=$(call if_deploying,/var/log/${PROJECT_PREFIX},${PWD}/${LOG_DIR})
CURRENT_HOST_LOCAL_DIR=$(call if_deploying,$(CURRENT_PROJECT_DIR)/${HOST_LOCAL_DIR},${PWD}/${HOST_LOCAL_DIR})

$(shell sed -i -e "/^\(#\|\)APP_HOST_NAME=/s&^.*$$&APP_HOST_NAME=$(CURRENT_APP_HOST_NAME)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)CERT_LOC=/s&^.*$$&CERT_LOC=$(CURRENT_CERT_LOC)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)CERT_KEY_LOC=/s&^.*$$&CERT_KEY_LOC=$(CURRENT_CERT_KEY_LOC)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)PROJECT_DIR=/s&^.*$$&PROJECT_DIR=$(CURRENT_PROJECT_DIR)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)LOG_DIR=/s&^.*$$&LOG_DIR=$(CURRENT_LOG_DIR)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)HOST_LOCAL_DIR=/s&^.*$$&HOST_LOCAL_DIR=$(CURRENT_HOST_LOCAL_DIR)&;" $(ENVFILE))

$(eval APP_HOST_NAME=$(CURRENT_APP_HOST_NAME))
$(eval CERT_LOC=$(CURRENT_CERT_LOC))
$(eval CERT_KEY_LOC=$(CURRENT_CERT_KEY_LOC))
$(eval PROJECT_DIR=$(CURRENT_PROJECT_DIR))
$(eval LOG_DIR=$(CURRENT_LOG_DIR))
$(eval HOST_LOCAL_DIR=$(CURRENT_HOST_LOCAL_DIR))

AI_ENABLED=$(shell [ $$(wc -c < ${AI_KEY_FILE}) -gt 5 ] && echo 1 || echo 0)

define clean_logs
  $(shell if [ $$(ls -1 $(LOG_DIR)/db/*.log 2>/dev/null | wc -l) -gt 1 ]; then \
    ls -t $(LOG_DIR)/db/*.log | tail -n +2 | xargs rm -f; \
  fi)
	rm -f log/*.log
endef

define clean_test
	@echo "======== cleaning tests ========"
	$(DOCKER_REDIS_EXEC) FLUSHALL
	$(call clean_logs)
endef

#################################
#             FLAGS             #
#################################

DOCKER_COMPOSE:=compose -f $(DOCKER_COMPOSE_SCRIPT) --env-file $(ENVFILE)
NO_LIMIT=-rateLimit=10000 -rateLimitBurst=10000
LOG_DEBUG=-logLevel=debug
LOG_NONE=-logLevel=
GO_ENVFILE_FLAG=GO_ENVFILE_LOC=${PROJECT_DIR}/.env
GO_DEV_FLAGS=$(GO_ENVFILE_FLAG)
GO_FUZZ_CACHEDIR=working/fuzz
GO_FUZZ_FLAGS=-test.fuzzcachedir=$(GO_FUZZ_CACHEDIR) -test.fuzztime=5s
GO_FUZZ_EXEC_FLAGS=-test.run=^$$ -test.fuzz=$${FUZZ:-.} -test.bench=^$$ -test.count=$${COUNT:-1} -test.v=$${V:-false} $(LOG_NONE) $(GO_FUZZ_FLAGS)
GO_TEST_EXEC_FLAGS=-test.run=$${TEST:-.} -test.fuzz=^$$ -test.bench=^$$ -test.count=$${COUNT:-1} -test.v=$${V:-false} $(LOG_NONE) $(GO_FUZZ_FLAGS)
GO_BENCH_EXEC_FLAGS=-test.run=^$$ -test.fuzz=^$$ -test.bench=$${BENCH:-.} -test.count=$${COUNT:-1} -test.v=$${V:-false} $(LOG_NONE) $(GO_FUZZ_FLAGS)
# $${PROF:-} # -cpuprofile=cpu.prof

GO=$(GO_ENVFILE_FLAG) go#GOEXPERIMENT=jsonv2 gotip# go

#################################
#             BUILDS            #
#################################

build: $(LOG_DIR) ${SIGNING_TOKEN_FILE} ${KC_PASS_FILE} ${KC_USER_CLIENT_SECRET_FILE} ${KC_API_CLIENT_SECRET_FILE} ${PG_PASS_FILE} ${PG_WORKER_PASS_FILE} ${REDIS_PASS_FILE} ${AI_KEY_FILE} $(CERT_LOC) $(CERT_KEY_LOC) $(JAVA_TARGET) $(LANDING_TARGET) $(TS_TARGET) $(TS_VAULT_WASM) $(PROTO_GEN_FILES) $(PROTO_GEN_MUTEX) $(PROTO_GEN_MUTEX_FILES) $(GO_HANDLERS_REGISTER) $(GO_TARGET)

# logs, certs, secrets, demo and backup dirs are not cleaned
.PHONY: clean
clean:
	rm -rf $(TS_BUILD_DIR) $(GO_GEN_DIR) \
		$(LANDING_BUILD_DIR) $(JAVA_TARGET_DIR) $(PLAYWRIGHT_CACHE_DIR)
	rm -f $(GO_TARGET) $(BINARY_TEST) $(TS_API_YAML) $(TS_API_BUILD) $(GO_HANDLERS_REGISTER) # $(MOCK_TARGET)

$(LOG_DIR):
ifeq ($(DEPLOYING),)
	mkdir -p $(LOG_DIR)/db
	setfacl -m g:$(H_GROUP):rwx $(LOG_DIR)
	setfacl -d -m g:$(H_GROUP):rwx $(LOG_DIR)
else
	sudo mkdir -p $(LOG_DIR)
	sudo setfacl -m g:$(H_GROUP):rwx $(LOG_DIR)
	sudo setfacl -d -m g:$(H_GROUP):rwx $(LOG_DIR)
	sudo mkdir -p $(LOG_DIR)/db
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

# ifeq ($(DEPLOYING),true)
#
# 	# sudo certbot certificates
# 	# sudo ip6tables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
# 	# sudo iptables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
# 	# sudo certbot certonly -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} \
# 	# 	--standalone --agree-tos --no-eff-email
# 	# sudo ip6tables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
# 	# sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
# 	# sudo chgrp -R ssl-certs /etc/letsencrypt/live /etc/letsencrypt/archive
# 	# sudo chmod -R g+r /etc/letsencrypt/live /etc/letsencrypt/archive
# 	# sudo chmod g+x /etc/letsencrypt/live /etc/letsencrypt/archive
# 	# # tailscale file cp .env "$(APP_HOST):"
# 	# # $(SSH) "sudo tailscale file get --conflict=overwrite $(H_ETC_DIR)/"


${SIGNING_TOKEN_FILE} ${KC_PASS_FILE} ${KC_USER_CLIENT_SECRET_FILE} ${KC_API_CLIENT_SECRET_FILE} ${PG_PASS_FILE} ${PG_WORKER_PASS_FILE} ${REDIS_PASS_FILE}:
	@mkdir -p -m 700 $(@D)
	openssl rand -hex 64 | tr -d '\n' > $@
	# chmod 600 $@
	# @mkdir -p $(@D)
	# install -m 640 /dev/null $@
	# openssl rand -hex 64 > $@ | tr -d '\n'
	# chmod 600 $@
	# setfacl -m u:${GID}:rw $@
# ifeq ($(DEPLOYING),true)
# 	@chgrp -R $(H_GROUP) $(@D)
# endif

${AI_KEY_FILE}:
	install -m 640 /dev/null $@
	echo "Provide an AiStudio API key if desired, or just press Enter"
	read -s AI_KEY; echo "$$AI_KEY" > $@
ifeq ($(DEPLOYING),true)
	@chgrp -R $(H_GROUP) $(@D)
endif

$(JAVA_TARGET): $(shell find $(JAVA_SRC)/{src,themes,pom.xml} -type f)
	rm -rf $(JAVA_SRC)/target
	mkdir $(@D)
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

$(TS_TARGET): $(TS_VAULT_WASM) $(TS_SRC)/.env.local $(TS_API_BUILD) $(shell find $(TS_SRC)/{src,public,package.json,index.html,vite.config.ts} -type f) $(shell find proto/ -type f)
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

$(TS_VAULT_WASM):
	cp $$(go env GOROOT)/lib/wasm/wasm_exec.js $(TS_PUBLIC_DIR)/wasm_exec.js
	GOOS=js GOARCH=wasm $(GO) build -C $(GO_VAULT_WASM_DIR) -ldflags="-s -w" -o $(abspath $(TS_VAULT_WASM)) ./...

$(GO_HANDLERS_REGISTER): $(GO_HANDLERS_REGISTER_CMD_DIR)/main.go $(PROTO_FILES)
	$(GO) run -C $(GO_HANDLERS_REGISTER_CMD_DIR) ./...

$(GO_TARGET): $(GO_FILES)
	$(GO) build -C $(GO_SRC) -o $(GO_TARGET) .

#################################
#           DEVELOP             #
#################################

.PHONY: go_dev
go_dev: 
	$(call clean_logs)
	$(GO_DEV_FLAGS) gow -e=go,mod run -C $(GO_SRC) . $(NO_LIMIT) $(LOG_DEBUG)

.PHONY: go_dev_ts
go_dev_ts: 
	$(call clean_logs)
	$(GO_DEV_FLAGS) gow -e=go,mod run -C $(GO_SRC) -tags=dev . $(NO_LIMIT) $(LOG_DEBUG)

.PHONY: go_tidy
go_tidy:
	cd $(GO_SRC) && go mod tidy

.PHONY: go_sec
go_sec:
	cd $(GO_SRC) && gosec -exclude-dir=types -exclude-dir=buf.build ./...

.PHONY: go_sec_all
go_sec_all:
	cd $(GO_SRC) && gosec ./...

.PHONY: go_watch
go_watch:
	$(call clean_logs)
	$(GO_DEV_FLAGS) gow -e=go,mod build -C $(GO_SRC) -o $(GO_TARGET) .

.PHONY: go_debug
go_debug:
	$(call clean_logs)
	cd go && $(GO_DEV_FLAGS) dlv debug --wd ${PWD}

.PHONY: go_debug_exec
go_debug_exec:
	$(call clean_logs)
	$(GO_DEV_FLAGS) dlv exec --wd ${PWD} $(GO_TARGET)

.PHONY: ts_dev
ts_dev:
	BROWSER=none HTTPS=true WDS_SOCKET_PORT=${GO_HTTPS_PORT} pnpm run --dir $(TS_SRC) start

.PHONY: landing_dev
landing_dev: build
	chmod +x $(LANDING_SRC)/server.sh && pnpm run --dir $(LANDING_SRC) start

#################################
#            TEST               #
#################################

.PHONY: go_test
go_test: go_test_integration go_test_bench

.PHONY: go_test_gen
go_test_gen:
	gotests -i -w -exported $(GO_SRC)
	gotests -i -w -exported $(GO_API_DIR)
	gotests -i -w -exported $(GO_CLIENTS_DIR)
	gotests -i -w -exported $(GO_HANDLERS_DIR)
	gotests -i -w -exported $(GO_UTIL_DIR)

.PHONY: go_coverage
go_coverage:
	$(GO) test -C $(GO_SRC) -coverpkg=./... ./...

.PHONY: go_test_gen_ui
test_gen:
	npx playwright codegen --ignore-https-errors https://localhost:${GO_HTTPS_PORT}

.PHONY: go_test_unit_api
go_test_unit_api:
	@$(call clean_test)
	$(GO) test -C $(GO_API_DIR) -c -o api.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_API_DIR)/api.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log
	@echo "api unit tests complete"

.PHONY: go_test_unit_clients
go_test_unit_clients:
	@$(call clean_test)
	$(GO) test -C $(GO_CLIENTS_DIR) -c -o clients.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_CLIENTS_DIR)/clients.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log
	@echo "clients unit tests complete"

.PHONY: go_test_unit_handlers
go_test_unit_handlers:
	@$(call clean_test)
	$(GO) test -C $(GO_HANDLERS_DIR) -c -o handlers.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_HANDLERS_DIR)/handlers.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log
	@echo "handlers unit tests complete"

.PHONY: go_test_unit_util
go_test_unit_util:
	@$(call clean_test)
	$(GO) test -C $(GO_UTIL_DIR) -c -o util.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_UTIL_DIR)/util.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log
	@echo "util unit tests complete"

.PHONY: go_test_unit
go_test_unit: $(GO_TARGET) go_test_unit_api go_test_unit_clients go_test_unit_handlers go_test_unit_util
	@echo "unit tests complete"

.PHONY: go_test_fuzz
go_test_fuzz: $(GO_TARGET)
	@mkdir -p $(GO_FUZZ_CACHEDIR)
	@$(call clean_test)
	$(GO) test -C $(GO_HANDLERS_DIR) -c -o handlers.fuzz.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_HANDLERS_DIR)/handlers.fuzz.$(BINARY_TEST) $(GO_FUZZ_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_ui
go_test_ui: $(GO_TARGET) docker_db_restore_op
	rm -f demos/*.webm
	$(call clean_test)
	$(GO) test -C $(GO_PLAYWRIGHT_DIR) -c -o playwright.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_PLAYWRIGHT_DIR)/playwright.$(BINARY_TEST) $(NO_LIMIT) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_integration
go_test_integration: $(GO_TARGET)
	$(call clean_test)
	$(GO) test -C $(GO_INTEGRATIONS_DIR) -short -c -o integration.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_INTEGRATIONS_DIR)/integration.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_integration_long
go_test_integration_long: $(GO_TARGET)
	$(call clean_test)
	$(GO) test -C $(GO_INTEGRATIONS_DIR) -c -o integration.long.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_INTEGRATIONS_DIR)/integration.long.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_integration_results
go_test_integration_results:
	less $(GO_INTEGRATIONS_DIR)/integration_results.json

.PHONY: go_test_bench_build
go_test_bench_build:
	$(GO) test -C $(GO_API_DIR) -c -o api.bench.$(BINARY_TEST) ./...
	$(GO) test -C $(GO_CLIENTS_DIR) -c -o clients.bench.$(BINARY_TEST) ./...
	$(GO) test -C $(GO_HANDLERS_DIR) -c -o handlers.bench.$(BINARY_TEST) ./...
	$(GO) test -C $(GO_UTIL_DIR) -c -o util.bench.$(BINARY_TEST) ./...

.PHONY: go_test_bench
go_test_bench: $(GO_TARGET) go_test_bench_build
	$(call clean_test)
	-$(GO_ENVFILE_FLAG) exec $(GO_API_DIR)/api.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)
	-$(GO_ENVFILE_FLAG) exec $(GO_CLIENTS_DIR)/clients.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)
	-$(GO_ENVFILE_FLAG) exec $(GO_HANDLERS_DIR)/handlers.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)
	-$(GO_ENVFILE_FLAG) exec $(GO_UTIL_DIR)/util.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)

#################################
#            DOCKER             #
#################################

.PHONY: docker_up
docker_up: build
	mkdir -p -m 700 ${UNIX_SOCK_DIR}
	install -d -m 700 -o ${UID} -g ${GID} $(UNIX_SOCK_DIR)/db $(UNIX_SOCK_DIR)/auth
	# setfacl -m u:${UID}:rwx $(UNIX_SOCK_DIR)/db
	# setfacl -d -m u:${UID}:rwx $(UNIX_SOCK_DIR)/db
	docker volume create $(PG_DATA) || true
	docker volume create $(REDIS_DATA) || true
	COMPOSE_BAKE=true docker $(DOCKER_COMPOSE) up -d --build
	chmod +x $(AUTH_INSTALL_SCRIPT) && exec $(AUTH_INSTALL_SCRIPT)
	
.PHONY: docker_down
docker_down:
	docker $(DOCKER_COMPOSE) down -v
	docker volume remove $(PG_DATA) || true
	docker volume remove $(REDIS_DATA) || true

.PHONY: docker_cycle
docker_cycle: docker_down docker_up docker_db_backup

.PHONY: docker_build
docker_build:
	docker $(DOCKER_COMPOSE) build

.PHONY: docker_start
docker_start: docker_build
	docker $(DOCKER_COMPOSE) up -d

.PHONY: docker_stop
docker_stop:
	docker $(DOCKER_COMPOSE) stop 

.PHONY: docker_db
docker_db:
	$(DOCKER_DB_EXEC) $(DOCKER_DB_CID) psql -U postgres -d ${PG_DB}

.PHONY: docker_redis
docker_redis:
	docker exec -it $(shell docker ps -aqf "name=redis") redis-cli --pass $$(cat $$REDIS_PASS_FILE)

#################################
#            HOST INIT          #
#################################

.PHONY: cloud_config_gen
cloud_config_gen:
	@mkdir -p $(HOST_LOCAL_DIR)
	date >> "$(HOST_LOCAL_DIR)/start_time"
	@echo "Tailscale auth key:"; \
	read -s TS_AUTH_KEY; \
	sed -e "s&dummyuser&${HOST_OPERATOR}&g; s&ts-auth-key&$$TS_AUTH_KEY&g; s&host-group&$(H_GROUP)&g; s&project-prefix&${PROJECT_PREFIX}&g; s&project-repo&https://${PROJECT_REPO}.git&g; s&https-port&${GO_HTTPS_PORT}&g; s&http-port&${GO_HTTP_PORT}&g; s&project-dir&$(H_ETC_DIR)&g; s&deploy-scripts&$(DEPLOY_HOST_SCRIPTS)&g; s&binary-name&${BINARY_NAME}&g; s&log-dir&/var/log/${PROJECT_PREFIX}&g; s&node-version&$(NODE_VERSION)&g; s&go-version&$(GO_VERSION)&g;" "$(DEPLOY_HOST_SCRIPTS)/cloud-config.yaml" > "$(HOST_LOCAL_DIR)/cloud-config.yaml"

.PHONY: host_local_up
host_local_up: cloud_config_gen
	> "$(HOST_LOCAL_DIR)"/.local
	multipass launch --name "${APP_HOST}" --timeout 900 --memory 4G --cpus 2 --disk 20G --cloud-init "$(HOST_LOCAL_DIR)/cloud-config.yaml"

.PHONY: host_local_down
host_local_down:
	multipass delete --purge "${APP_HOST}"
	rm -rf "$(HOST_LOCAL_DIR)"

.PHONY: host_up
host_up: cloud_config_gen
	cp "$(DEPLOY_HOST_SCRIPTS)/public-firewall.json" "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud firewall create --name "${PROJECT_PREFIX}-public-firewall" --rules-file "$(HOST_LOCAL_DIR)/public-firewall.json" >/dev/null
	hcloud firewall describe "${PROJECT_PREFIX}-public-firewall" -o json > "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud server create \
		--name "${APP_HOST}" --datacenter "${HETZNER_DATACENTER}" \
		--type "${HETZNER_APP_TYPE}" --image "${HETZNER_IMAGE}" \
		--user-data-from-file "$(HOST_LOCAL_DIR)/cloud-config.yaml" --firewall "${PROJECT_PREFIX}-public-firewall" >/dev/null
	hcloud server describe "${APP_HOST}" -o json > "$(HOST_LOCAL_DIR)/app.json"
	jq -r '.public_net.ipv6.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip6"
	jq -r '.public_net.ipv4.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip"
	@echo "Server is deploying! This will take a moment..."
	@echo "Continue with make host_install_service, once ${APP_HOST} appears in Tailscale and some time is allowed for server reboot."

# ,
#   {
#     "direction": "in",
#     "protocol": "udp",
#     "port": "44400-44500",
#     "source_ips": ["0.0.0.0/0", "::/0"]
#   }

.PHONY: host_install_service
host_install_service: host_sync_files
	if [ ! -f "$(HOST_LOCAL_DIR)/.local" ]; then \
		$(MAKE) host_update_cert; \
	else \
		$(MAKE) host_local_update_cert; \
	fi
	$(SSH) "cd $(H_ETC_DIR) && make host_install_service_op"

.PHONY: host_install_service_op
host_install_service_op:
	sudo mkdir -p /etc/logrotate.d/${PROJECT_PREFIX}
	sed -e 's&project-dir&$(H_ETC_DIR)&g;' "$(CRON_SCRIPTS)/whitelist-ips" | sudo tee "/etc/cron.daily/whitelist-ips" >/dev/null
	sudo chmod 755 "/etc/cron.daily/whitelist-ips"
	sed -e 's&project-prefix&${PROJECT_PREFIX}&g;' "$(DEPLOY_HOST_SCRIPTS)/jail.local" | sudo tee /etc/fail2ban/jail.local >/dev/null
	sed -e 's&project-prefix&${PROJECT_PREFIX}&g;' "$(DEPLOY_HOST_SCRIPTS)/logrotate.conf" | sudo tee /etc/logrotate.d/${PROJECT_PREFIX}/logrotate.conf >/dev/null
	sudo cp "$(DEPLOY_HOST_SCRIPTS)/http-auth.conf" "$(DEPLOY_HOST_SCRIPTS)/http-access.conf" /etc/fail2ban/filter.d/
	sudo cp "$(DEPLOY_HOST_SCRIPTS)/ufw-subnet.conf" /etc/fail2ban/action.d/
	sed -e 's&binary-name&${BINARY_NAME}&g; s&etc-dir&$(H_ETC_DIR)&g' "$(DEPLOY_HOST_SCRIPTS)/start.sh" > start.sh
	sed -e 's&host-operator&${HOST_OPERATOR}&g; s&host-group&${H_GROUP}&g; s&etc-dir&$(H_ETC_DIR)&g' "$(DEPLOY_HOST_SCRIPTS)/host.service" > $(BINARY_SERVICE)
	sudo install -m 750 -o ${HOST_OPERATOR} -g ${HOST_OPERATOR} start.sh /usr/local/bin
	sudo install -m 644 $(BINARY_SERVICE) /etc/systemd/system
	rm start.sh $(BINARY_SERVICE)
	sudo systemctl restart fail2ban
	sudo systemctl enable $(BINARY_SERVICE)

.PHONY: host_sync_files
host_sync_files:
	tailscale file cp .env "$(APP_HOST):"
	$(SSH) " \
		sudo tailscale file get --conflict=overwrite $(H_ETC_DIR)/; \
		sudo chown $(H_LOGIN):$(H_GROUP) $(H_ETC_DIR)/.env; \
	"
	if [ -f "$(DEMOS_DIR)" ]; then \
		tailscale file cp "$(DEMOS_DIR)/"* "$(APP_HOST):"; \
		$(SSH) " \
			sudo tailscale file get --conflict=overwrite $(H_ETC_DIR)/$(DEMOS_DIR)/; \
			sudo chown -R $(H_LOGIN):$(H_GROUP) $(H_ETC_DIR)/$(DEMOS_DIR)/; \
		"; \
	fi

.PHONY: host_local_update_cert
host_local_update_cert:
	$(SSH) " \
		sudo mkdir -p /etc/letsencrypt/live/${DOMAIN_NAME}; \
		sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
			-keyout /etc/letsencrypt/live/${DOMAIN_NAME}/privkey.pem \
			-out /etc/letsencrypt/live/${DOMAIN_NAME}/cert.pem \
			-subj \"/C=US/ST=State/L=City/O=Organization/CN=${DOMAIN_NAME}\"; \
	"
	$(SSH) " \
		sudo chgrp -R ssl-certs /etc/letsencrypt/live; \
		sudo chmod -R g+r /etc/letsencrypt/live; \
		sudo chmod g+x /etc/letsencrypt/live; \
	"

# if we don't have certs locally or we're renewing, do the normal cert request and store the certs locally
# if the server still doesn't have certs then we aren't renewing and already have certs locally, likely new deployment
.PHONY: host_update_cert
host_update_cert: host_install_cert host_replace_cert
	@echo "certs handled"

.PHONY: host_install_cert
host_install_cert:
	@if [ ! -d "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}" ] || [ -n "$$RENEW_CERT" ]; then \
		$(SSH) " \
			sudo rm -rf /etc/letsencrypt/archive /etc/letsencrypt/live; \
			sudo iptables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT} || true; \
			sudo certbot certonly --standalone -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} --agree-tos --no-eff-email; \
			sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
		"; \
		$(MAKE) host_group_cert; \
		$(MAKE) host_backup_cert; \
		echo "installed certs, renew=$$RENEW_CERT"; \
	else \
		echo "skipping install"; \
	fi

.PHONY: host_replace_cert
host_replace_cert:
	@if $(SSH) "[ ! -d /etc/letsencrypt/archive/${DOMAIN_NAME} ]"; then \
		tailscale file cp "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}/"* "$(APP_HOST):"; \
		$(SSH) " \
			sudo mkdir -p /etc/letsencrypt/archive/${DOMAIN_NAME}/ /etc/letsencrypt/live/${DOMAIN_NAME}/; \
			sudo tailscale file get --conflict=overwrite /etc/letsencrypt/archive/${DOMAIN_NAME}/; \
			sudo find /etc/letsencrypt/archive/${DOMAIN_NAME} -maxdepth 1 -type f | while read file; \
			do sudo ln -s \"\$$file\" /etc/letsencrypt/live/${DOMAIN_NAME}/\$$(basename \"\$$file\"); \
			done \
		"; \
		$(MAKE) host_group_cert; \
		echo "replaced certs into new directories"; \
	else \
		echo "using existing certs"; \
	fi

.PHONY: host_group_cert
host_group_cert:
	$(SSH) " \
		sudo chgrp -R ssl-certs /etc/letsencrypt/live /etc/letsencrypt/archive; \
		sudo chmod -R g+r /etc/letsencrypt/live /etc/letsencrypt/archive; \
		sudo chmod g+x /etc/letsencrypt/live /etc/letsencrypt/archive; \
	"

.PHONY: host_backup_cert
host_backup_cert:
	$(SSH) " \
		mkdir -p \"$(H_ETC_DIR)/$(CERT_BACKUP_DIR)\"; \
		sudo cp -a /etc/letsencrypt/archive/${DOMAIN_NAME}/* \"$(H_ETC_DIR)/$(CERT_BACKUP_DIR)\"; \
		sudo tailscale file cp \"$(H_ETC_DIR)/$(CERT_BACKUP_DIR)/\"* $(shell hostname):; \
	"
	mkdir -p "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}"
	tailscale file get --conflict=overwrite "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}/"

#################################
#           HOST UTILS          #
#################################

.PHONY: host_down
host_down:
	hcloud server delete "${APP_HOST}"
	hcloud firewall delete "${PROJECT_PREFIX}-public-firewall"
	rm -rf $(HOST_LOCAL_DIR)

.PHONY: host_reboot
host_reboot:
	hcloud server reboot "${APP_HOST}"
	until ping -c1 "${APP_HOST}" ; do sleep 5; done
	@echo "rebooted"

.PHONY: host_update
host_update: $(LANDING_SRC)/config.yaml
	git reset --hard HEAD
	git pull
	sed -i -e '/^  lastUpdated:/s/^.*$$/  lastUpdated: $(shell date +%Y-%m-%d)/' $(LANDING_SRC)/config.yaml

.PHONY: host_deploy
host_deploy: host_sync_files 
	$(SSH) "cd $(H_ETC_DIR) && make host_update && make build && make host_service_start_op"

.PHONY: host_service_start
host_service_start:
	$(SSH) "cd $(H_ETC_DIR) && make host_service_start_op"

.PHONY: host_service_start_op
host_service_start_op: docker_up
	sudo install -m 700 -o ${HOST_OPERATOR} -g 1000 $(GO_TARGET) /usr/local/bin
	sudo systemctl restart $(BINARY_SERVICE)
	sudo systemctl is-active $(BINARY_SERVICE)

.PHONY: host_service_stop
host_service_stop:
	$(SSH) "cd $(H_ETC_DIR) && make host_service_stop_op"

.PHONY: host_service_stop_op
host_service_stop_op:
	sudo systemctl stop $(BINARY_SERVICE)

.PHONY: host_ssh
host_ssh:
	$(SSH)

.PHONY: host_status
host_status:
	$(SSH) sudo journalctl -n 100 -u ${BINARY_SERVICE} -f

.PHONY: host_errors
host_errors:
	$(SSH) "tail -n 100 -f $(LOG_DIR)/errors.log"

.PHONY: host_db
host_db:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="db") psql -U postgres ${PG_DB}

.PHONY: host_redis
host_redis:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="redis") redis-cli --pass ${REDIS_PASS}

.PHONY: host_metric_cpu
host_metric_cpu:
	hcloud server metrics --type cpu $(APP_HOST)

.PHONY: host_run_cron
host_run_cron:
	$(SSH) 'run-parts /etc/cron.daily'

#################################
#            BACKUP             #
#################################

.PHONY: docker_db_redeploy
docker_db_redeploy: docker_stop
	docker rm $(DOCKER_DB_CID) || true
	docker volume remove $(PG_DATA) || true
	docker volume create $(PG_DATA)
	COMPOSE_BAKE=true docker $(DOCKER_COMPOSE) up -d --build db
	sleep 5

.PHONY: docker_db_backup
docker_db_backup:
	mkdir -p $(DB_BACKUP_DIR)
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_dump --inserts --on-conflict-do-nothing -Fc keycloak \
		> $(DB_BACKUP_DIR)/${PG_DB}_keycloak_$(shell TZ=UTC date +%Y%m%d%H%M%S).dump
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_dump --column-inserts --data-only --on-conflict-do-nothing -n dbtable_schema -Fc ${PG_DB} \
		> $(DB_BACKUP_DIR)/${PG_DB}_app_$(shell TZ=UTC date +%Y%m%d%H%M%S).dump

.PHONY: docker_db_restore_op
docker_db_restore_op:
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_restore -c -d keycloak \
		< $(DB_BACKUP_DIR)/$(shell ls -Art --ignore "${PG_DB}_app*" $(DB_BACKUP_DIR) | tail -n 1) || true
	docker exec -i $(shell docker ps -aqf "name=db") psql -U postgres -d ${PG_DB} -c " \
		TRUNCATE TABLE dbtable_schema.users CASCADE; \
		TRUNCATE TABLE dbtable_schema.roles CASCADE; \
		TRUNCATE TABLE dbtable_schema.file_types CASCADE; \
		TRUNCATE TABLE dbtable_schema.budgets CASCADE; \
		TRUNCATE TABLE dbtable_schema.timelines CASCADE; \
		TRUNCATE TABLE dbtable_schema.time_units CASCADE; \
	"
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_restore -a --disable-triggers --superuser=postgres -d ${PG_DB} \
		< $(DB_BACKUP_DIR)/$(shell ls -Art --ignore "${PG_DB}_keycloak*" $(DB_BACKUP_DIR) | tail -n 1) || true

.PHONY: docker_db_restore
docker_db_restore: docker_db_redeploy docker_db_restore_op docker_start

.PHONY: docker_db_redeploy_views
docker_db_redeploy_views:
	@> working/sql_views
	@printf "DROP SCHEMA dbview_schema CASCADE;\nCREATE SCHEMA dbview_schema;\nGRANT USAGE ON SCHEMA dbview_schema TO $(PG_WORKER);\nALTER DEFAULT PRIVILEGES IN SCHEMA dbview_schema GRANT ALL ON TABLES TO $(PG_WORKER);\n\n" >> working/sql_views
	@cat $(DB_SCRIPTS)/base_views.sql >> working/sql_views
	@echo "" >> working/sql_views
	@cat $(DB_SCRIPTS)/app_views.sql >> working/sql_views
	@echo "" >> working/sql_views
	@cat $(DB_SCRIPTS)/function_views.sql >> working/sql_views
	@echo "" >> working/sql_views
	@cat $(DB_SCRIPTS)/kiosk_views.sql >> working/sql_views
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) psql -U postgres -d ${PG_DB} < working/sql_views

.PHONY: host_db_backup
host_db_backup:
	@mkdir -p "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"
	$(SSH) 'cd $(H_ETC_DIR) && make docker_db_backup && tailscale file cp "$(H_ETC_DIR)/$(DB_BACKUP_DIR)/"* "$(shell hostname)"'
	sudo tailscale file get "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"

.PHONY: host_db_backup_restore
host_db_backup_restore:
	tailscale file cp "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"* "$(APP_HOST):"
	$(SSH) "sudo tailscale file get $(H_ETC_DIR)/$(DB_BACKUP_DIR)/"

.PHONY: host_db_restore_op
host_db_restore_op:
	$(SSH) "cd $(H_ETC_DIR) && make docker_db_restore"

.PHONY: host_db_restore
host_db_restore: host_service_stop host_db_restore_op host_service_start

.PHONY: check_logs
check_logs:
	exec $(CRON_SCRIPTS)/check-logs

