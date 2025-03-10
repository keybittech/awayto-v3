SHELL := /bin/bash

GO_VERSION=go1.24.0.linux-amd64
NODE_VERSION=v22.13.1

# manually manage path for makefile use
export PATH := ${PATH}:/home/$(shell whoami)/.nvm/versions/node/$(NODE_VERSION)/bin:/home/$(shell whoami)/go/bin:/usr/local/go/bin

ENVFILE?=./.env

$(shell if [ ! -f $(ENVFILE) ]; then install -m 600 .env.template $(ENVFILE); fi)

include $(ENVFILE)
export $(shell sed 's/=.*//' $(ENVFILE))

# go build output

BINARY_TEST=$(BINARY_NAME).test
BINARY_SERVICE=$(BINARY_NAME).service

# source files

JAVA_SRC=./java
TS_SRC=./ts
GO_SRC=./go
LANDING_SRC=./landing

# local directories

JAVA_TARGET_DIR=java/target
JAVA_THEMES_DIR=java/themes
LANDING_BUILD_DIR=landing/public
TS_BUILD_DIR=ts/build
HOST_LOCAL_DIR=sites/${PROJECT_PREFIX}
GO_GEN_DIR=go/pkg/types
GO_MOCKS_GEN_DIR=go/pkg/mocks
export PLAYWRIGHT_CACHE_DIR=working/playwright # export here for test runner to see
DEMOS_DIR=demos/final

# build targets

JAVA_TARGET=$(JAVA_TARGET_DIR)/custom-event-listener.jar
LANDING_TARGET=$(LANDING_BUILD_DIR)/index.html
TS_TARGET=$(TS_BUILD_DIR)/index.html
GO_TARGET=${PWD}/go/$(BINARY_NAME)
GO_MOCK_TARGET=$(GO_MOCKS_GEN_DIR)/clients.go
PROTO_MOD_TARGET=$(GO_GEN_DIR)/go.mod

# host locations

H_OP=/home/${HOST_OPERATOR}
H_DOCK=$(H_OP)/bin/docker
H_REM_DIR=$(H_OP)/${PROJECT_PREFIX}
H_ETC_DIR=/etc/${PROJECT_PREFIX}

H_SIGN=${HOST_OPERATOR}@$$(cat "$(HOST_LOCAL_DIR)/app_ip")

# build artifacts

PROTO_FILES=$(wildcard proto/*.proto)

TS_API_YAML=ts/openapi.yaml
TS_API_BUILD=ts/src/hooks/api.ts
TS_CONFIG_API=ts/openapi-config.json

CLOUD_CONFIG_OUTPUT=$(HOST_LOCAL_DIR)/cloud-config.yaml

# deploy properties
RSYNC_FLAGS=-ave 'ssh -p ${SSH_PORT}'
DOCKER_COMPOSE:=compose -f deploy/scripts/docker/docker-compose.yml --env-file $(ENVFILE)

DEPLOY_SCRIPTS=deploy/scripts
DEV_SCRIPTS=deploy/scripts/dev
DOCKER_SCRIPTS=deploy/scripts/docker
DEPLOY_HOST_SCRIPTS=deploy/scripts/host
CRON_DIR=deploy/scripts/cron
DEPLOY_SCRIPT=deploy/scripts/host/deploy.sh
AUTH_INSTALL_SCRIPT=deploy/scripts/auth/install.sh

# backup related
DB_BACKUP_DIR=backups/db
DOCKER_DB_CID=$(shell ${SUDO} docker ps -aqf "name=db")
DOCKER_DB_EXEC := ${SUDO} docker exec --user postgres -it
DOCKER_DB_CMD := ${SUDO} docker exec --user postgres -i


# flags
GO_DEV_FLAGS=GO_ENVFILE_LOC=${PWD}/.env GO_ERROR_LOG=${GO_ERROR_LOG} SIGNING_TOKEN_FILE=${SIGNING_TOKEN_FILE} LOG_LEVEL=debug
SSH=ssh -p ${SSH_PORT} -T $(H_SIGN)

LOCAL_UNIX_SOCK_DIR=$(shell pwd)/${UNIX_SOCK_DIR_NAME}
define set_local_unix_sock_dir
	$(eval UNIX_SOCK_DIR=${LOCAL_UNIX_SOCK_DIR})
endef

CURRENT_USER:=$(shell whoami)
DEPLOYING:=$(if $(filter ${HOST_OPERATOR},${CURRENT_USER}),true,)

define if_deploying
$(if $(DEPLOYING),$(1),$(2))
endef

CURRENT_APP_HOST_NAME=$(call if_deploying,${DOMAIN_NAME},localhost:${GO_HTTPS_PORT})
CURRENT_CERTS_DIR=$(call if_deploying,/etc/letsencrypt/live/${DOMAIN_NAME},${CERTS_DIR})
CURRENT_PROJECT_DIR=$(call if_deploying,/home/${HOST_OPERATOR}/${PROJECT_PREFIX},${PWD})

$(shell sed -i -e "/^\(#\|\)APP_HOST_NAME=/s&^.*$$&APP_HOST_NAME=$(CURRENT_APP_HOST_NAME)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)CERTS_DIR=/s&^.*$$&CERTS_DIR=$(CURRENT_CERTS_DIR)&;" $(ENVFILE))
$(shell sed -i -e "/^\(#\|\)PROJECT_DIR=/s&^.*$$&PROJECT_DIR=$(CURRENT_PROJECT_DIR)&;" $(ENVFILE))

$(eval APP_HOST_NAME=$(CURRENT_APP_HOST_NAME))
$(eval CERTS_DIR=$(CURRENT_CERTS_DIR))
$(eval PROJECT_DIR=$(CURRENT_PROJECT_DIR))

AI_ENABLED=$(shell [ $$(wc -c < ${OAI_KEY_FILE}) -gt 1 ] && echo 1 || echo 0)

#################################
#           TARGETS             #
#################################

build: ${SIGNING_TOKEN_FILE} ${KC_PASS_FILE} ${KC_API_CLIENT_SECRET_FILE} ${PG_PASS_FILE} ${PG_WORKER_PASS_FILE} ${REDIS_PASS_FILE} ${OAI_KEY_FILE} ${CERT_LOC} ${CERT_KEY_LOC} $(JAVA_TARGET) $(LANDING_TARGET) $(TS_TARGET) $(GO_TARGET)

# certs, secrets, demo and backup dirs are not cleaned
.PHONY: clean
clean:
	rm -rf $(TS_BUILD_DIR) $(GO_MOCKS_GEN_DIR) $(GO_GEN_DIR) \
		$(LANDING_BUILD_DIR) $(JAVA_TARGET_DIR) $(PLAYWRIGHT_CACHE_DIR)
	rm -f $(GO_TARGET) $(BINARY_TEST) $(TS_API_YAML) $(TS_API_BUILD)

${CERT_LOC} ${CERT_KEY_LOC}:
ifeq ($(DEPLOYING),true)
	sudo certbot certificates
	sudo ip6tables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo iptables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo certbot certonly -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} \
		--standalone --agree-tos --no-eff-email
	sudo ip6tables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} /etc/letsencrypt
	sudo chmod 600 ${CERT_LOC} ${CERT_KEY_LOC}
else
	mkdir -p $(@D)
	chmod 750 $(@D)
	openssl req -nodes -new -x509 -keyout ${CERT_KEY_LOC} -out ${CERT_LOC} -days 365 -subj "/CN=${APP_HOST_NAME}"
	chmod 600 ${CERT_KEY_LOC} ${CERT_LOC}
	echo "pwd required to update cert chain"
	sudo cp ${CERT_LOC} /usr/local/share/ca-certificates
	sudo update-ca-certificates
endif

${SIGNING_TOKEN_FILE} ${KC_PASS_FILE} ${KC_API_CLIENT_SECRET_FILE} ${PG_PASS_FILE} ${PG_WORKER_PASS_FILE} ${REDIS_PASS_FILE}:
	@mkdir -p $(@D)
	@chmod 750 $(@D)
	install -m 640 /dev/null $@
	openssl rand -hex 64 > $@ | tr -d '\n'

${OAI_KEY_FILE}:
	install -m 640 /dev/null $@
	if [[ ! -v OPENAI_API_KEY ]]; then \
		echo "Provide an OPENAI_API_KEY if desired, or just press Enter"; \
		read -s OAI_KEY; echo "$$OAI_KEY" > $@ ;\
	else \
		echo "$$OPENAI_API_KEY" > $@; \
	fi

$(JAVA_TARGET): $(shell find $(JAVA_SRC)/{src,themes,pom.xml} -type f)
	rm -rf $(JAVA_SRC)/target
	mvn -f $(JAVA_SRC) install

$(LANDING_SRC)/config.yaml: $(LANDING_SRC)/config.yaml.template
	sed -e 's&project-title&${PROJECT_TITLE}&g; s&last-updated&$(shell date +%Y-%m-%d)&g; s&app-host-url&${APP_HOST_URL}&g;' "$(LANDING_SRC)/config.yaml.template" > "$(LANDING_SRC)/config.yaml"

# using npm here as pnpm symlinks just hugo and doesn't build correctly 
$(LANDING_TARGET): $(LANDING_SRC)/config.yaml $(shell find $(LANDING_SRC)/{assets,content,layouts,static,package-lock.json} -type f)
	npm --prefix ${LANDING_SRC} i
	npm run --prefix ${LANDING_SRC} build

$(TS_API_BUILD): $(shell find proto/ -type f)
	protoc --proto_path=proto \
		--experimental_allow_proto3_optional \
		--openapi_out=$(TS_SRC) \
		$(PROTO_FILES)
	npx -y @rtk-query/codegen-openapi $(TS_CONFIG_API)

$(TS_SRC)/.env.local: $(TS_SRC)/.env.template
	sed -e 's&app-host-url&${APP_HOST_URL}&g; s&app-host-name&${APP_HOST_NAME}&g; s&kc-realm&${KC_REALM}&g; s&kc-client&${KC_CLIENT}&g; s&kc-path&${KC_PATH}&g; s&turn-name&${TURN_NAME}&g; s&turn-pass&${TURN_PASS}&g; s&allowed-file-ext&${ALLOWED_FILE_EXT}&g; s&ai-enabled&$(AI_ENABLED)&g;' "$(TS_SRC)/.env.template" > "$(TS_SRC)/.env.local"

$(TS_TARGET): $(TS_SRC)/.env.local $(TS_API_BUILD) $(shell find $(TS_SRC)/{src,public,package.json,index.html,vite.config.ts} -type f) $(shell find proto/ -type f)
	pnpm --dir $(TS_SRC) i
	pnpm run --dir $(TS_SRC) build

$(PROTO_MOD_TARGET): $(shell find proto/ -type f)
	@mkdir -p $(@D)
	protoc --proto_path=proto \
		--experimental_allow_proto3_optional \
		--go_out=$(GO_GEN_DIR) \
		--go_opt=module=${PROJECT_REPO}/$(GO_GEN_DIR) \
		$(PROTO_FILES)
	if [ ! -f "$(GO_GEN_DIR)/go.mod" ]; then \
		cd $(@D) && go mod init ${PROJECT_REPO}/$(GO_GEN_DIR) && go mod tidy && cd -; \
	fi

$(GO_TARGET): $(shell find $(GO_SRC)/{main.go,pkg} -type f) $(PROTO_MOD_TARGET) $(GO_MOCK_TARGET)
	$(call set_local_unix_sock_dir)
	cd go && go mod tidy && cd -
	go build -C $(GO_SRC) -o $(GO_TARGET) .

$(GO_MOCK_TARGET): 
	@mkdir -p $(@D)
	mockgen -source=go/pkg/clients/interfaces.go -destination=$(GO_MOCK_TARGET) -package=mocks
	cd $(@D) && \
		go mod init ${PROJECT_REPO}/go/pkg/mocks && \
		go mod edit -replace ${PROJECT_REPO}/go/pkg/api=../api && \
		go mod edit -replace ${PROJECT_REPO}/go/pkg/clients=../clients && \
		go mod edit -replace ${PROJECT_REPO}/go/pkg/handlers=../handlers && \
		go mod edit -replace ${PROJECT_REPO}/go/pkg/types=../types && \
		go mod edit -replace ${PROJECT_REPO}/go/pkg/util=../util && \
		go mod tidy && \
		cd -

#################################
#           DEVELOP             #
#################################

.PHONY: go_dev
go_dev:
	$(call set_local_unix_sock_dir)
	$(GO_DEV_FLAGS) gow -e=go,mod run -C $(GO_SRC) .

.PHONY: go_watch
go_watch:
	$(call set_local_unix_sock_dir)
	$(GO_DEV_FLAGS) gow -e=go,mod build -C $(GO_SRC) -o $(GO_TARGET) .

.PHONY: go_debug
go_debug:
	$(call set_local_unix_sock_dir)
	cd go && $(GO_DEV_FLAGS) dlv debug --wd ${PWD}

.PHONY: go_debug_exec
go_debug_exec:
	$(call set_local_unix_sock_dir)
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
go_test: docker_up go_test_main go_test_pkg

.PHONY: go_test_main
go_test_main: build $(GO_TARGET)
	$(call set_local_unix_sock_dir)
	go test -C $(GO_SRC) -v -c -o ../$(BINARY_TEST) && exec ./$(BINARY_TEST)

.PHONY: go_test_pkg
go_test_pkg: $(GO_TARGET) $(GO_MOCK_TARGET)
	$(call set_local_unix_sock_dir)
	go test -C $(GO_SRC) -v ./...

.PHONY: go_coverage
go_coverage: $(GO_MOCK_TARGET)
	go test -C $(GO_SRC) -coverpkg=./... ./...

.PHONY: test_clean
test_clean:
	rm -rf $(PLAYWRIGHT_CACHE_DIR)

.PHONY: test_gen
test_gen:
	npx playwright codegen --ignore-https-errors https://localhost:${GO_HTTPS_PORT}

#################################
#            DOCKER             #
#################################

.PHONY: docker_up
docker_up:
	$(call set_local_unix_sock_dir)
	${SUDO} docker volume create $(PG_DATA)
	${SUDO} docker volume create $(REDIS_DATA)
	COMPOSE_BAKE=true ${SUDO} docker $(DOCKER_COMPOSE) up -d --build
	chmod +x $(AUTH_INSTALL_SCRIPT) && exec $(AUTH_INSTALL_SCRIPT)

.PHONY: docker_down
docker_down:
	${SUDO} docker $(DOCKER_COMPOSE) down 
	${SUDO} docker volume remove $(PG_DATA) || true
	${SUDO} docker volume remove $(REDIS_DATA) || true

.PHONY: docker_cycle
docker_cycle: docker_down docker_up

.PHONY: docker_build
docker_build:
	$(call set_local_unix_sock_dir)
	${SUDO} docker $(DOCKER_COMPOSE) build

.PHONY: docker_start
docker_start: docker_build
	${SUDO} docker $(DOCKER_COMPOSE) up -d
	chmod +x $(AUTH_INSTALL_SCRIPT) && exec $(AUTH_INSTALL_SCRIPT)

.PHONY: docker_stop
docker_stop:
	${SUDO} docker $(DOCKER_COMPOSE) stop 

.PHONY: docker_db
docker_db:
	$(DOCKER_DB_EXEC) $(DOCKER_DB_CID) psql -U postgres -d ${PG_DB}

.PHONY: docker_redis
docker_redis:
	${SUDO} docker exec -it $(shell docker ps -aqf "name=redis") redis-cli --pass $$(cat $$REDIS_PASS_FILE)

#################################
#            HOST INIT          #
#################################

.PHONY: host_up
host_up: 
	@mkdir -p $(HOST_LOCAL_DIR)
	date >> "$(HOST_LOCAL_DIR)/start_time"
	@sed -e 's&dummyuser&${HOST_OPERATOR}&g; s&id-rsa-pub&$(shell cat ${RSA_PUB})&g; s&project-prefix&${PROJECT_PREFIX}&g; s&ssh-port&${SSH_PORT}&g; s&project-repo&https://${PROJECT_REPO}.git&g; s&https-port&${GO_HTTPS_PORT}&g; s&http-port&${GO_HTTP_PORT}&g;' "$(DEPLOY_HOST_SCRIPTS)/cloud-config.yaml" > "$(HOST_LOCAL_DIR)/cloud-config.yaml"
	sed -e 's&ssh-port&${SSH_PORT}&g;' "$(DEPLOY_HOST_SCRIPTS)/public-firewall.json" > "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud firewall create --name "${PROJECT_PREFIX}-public-firewall" --rules-file "$(HOST_LOCAL_DIR)/public-firewall.json" >/dev/null
	hcloud firewall describe "${PROJECT_PREFIX}-public-firewall" -o json > "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud server create \
		--name "${APP_HOST}" --datacenter "${HETZNER_DATACENTER}" \
		--type "${HETZNER_APP_TYPE}" --image "${HETZNER_IMAGE}" \
		--user-data-from-file "$(HOST_LOCAL_DIR)/cloud-config.yaml" --firewall "${PROJECT_PREFIX}-public-firewall" >/dev/null
	hcloud server describe "${APP_HOST}" -o json > "$(HOST_LOCAL_DIR)/app.json"
	jq -r '.public_net.ipv6.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip6"
	jq -r '.public_net.ipv4.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip"
	until ssh-keyscan -p ${SSH_PORT} -H $$(cat "$(HOST_LOCAL_DIR)/app_ip") >> ~/.ssh/known_hosts; do sleep 5; done
	$(SSH) sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} $(H_REM_DIR)
	make host_sync_env
	$(SSH) 'cd "$(H_REM_DIR)" && make host_install'

.PHONY: host_install
host_install:
	sudo ip6tables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo ip6tables -A PREROUTING -t nat -p tcp --dport 443 -j REDIRECT --to-port ${GO_HTTPS_PORT}
	sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo iptables -A PREROUTING -t nat -p tcp --dport 443 -j REDIRECT --to-port ${GO_HTTPS_PORT}
	mkdir -p $(H_REM_DIR)/{backups/db,working,local_tmp,demos/final}
	@echo "installing nvm"
	curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
	. ~/.nvm/nvm.sh && nvm install $(NODE_VERSION) && npm i -g pnpm@latest-10
	@echo "installing go"
	sudo rm -rf /usr/local/go
	sudo curl -L -o goinstall.tar.gz https://go.dev/dl/$(GO_VERSION).tar.gz
	sudo tar -C /usr/local -xzf goinstall.tar.gz
	if ! grep -q "go/bin" "$(H_OP)/.bashrc"; then \
		echo "export PATH=\$$PATH:/usr/local/go/bin" >> $(H_OP)/.bashrc; \
		echo "clear && cd $(H_REM_DIR)" >> $(H_OP)/.bashrc; \
	fi
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/golang/mock/mockgen@v1.6.0
	sudo tailscale up

.PHONY: host_reboot
host_reboot:
	hcloud server reboot "${APP_HOST}"
	until ping -c1 "${APP_HOST}" ; do sleep 5; done
	@echo "rebooted"

.PHONY: host_down
host_down:
	ssh-keygen -f ~/.ssh/known_hosts -R "${APP_HOST}:${SSH_PORT}"
	hcloud server delete "${APP_HOST}"
	hcloud firewall delete "${PROJECT_PREFIX}-public-firewall"
	rm -rf $(HOST_LOCAL_DIR)

.PHONY: host_sync_env
host_sync_env:
	mkdir -p $(HOST_LOCAL_DIR)/cron/daily
	@sed -e 's&dummyuser&${HOST_OPERATOR}&g; s&project-prefix&${PROJECT_PREFIX}&g;' "$(CRON_DIR)/whitelist-ips" > "$(HOST_LOCAL_DIR)/cron/daily/whitelist-ips"
	rsync ${RSYNC_FLAGS} --chown root:root --chmod 755 --rsync-path="sudo rsync" "$(HOST_LOCAL_DIR)/cron/daily/" "$(H_SIGN):/etc/cron.daily/"
	$(SSH) 'run-parts /etc/cron.daily'
	rsync ${RSYNC_FLAGS} "$(DEMOS_DIR)/" "$(H_SIGN):$(H_REM_DIR)/$(DEMOS_DIR)/"
	rsync ${RSYNC_FLAGS} --chown ${HOST_OPERATOR}:${HOST_OPERATOR} --chmod 400 .env "$(H_SIGN):$(H_REM_DIR)"

#################################
#           HOST UTILS          #
#################################

.PHONY: host_deploy
host_deploy: host_sync_env
	$(SSH) 'cd "$(H_REM_DIR)" && sh $(DEPLOY_SCRIPT)'

.PHONY: host_update_cert
host_update_cert:
	$(SSH) 'cd "$(H_REM_DIR)" && make host_update_cert_op'

.PHONY: host_ssh
host_ssh:
	@ssh -p ${SSH_PORT} $(H_SIGN)

.PHONY: host_status
host_status:
	$(SSH) sudo journalctl -n 100 -u ${BINARY_SERVICE} -f

.PHONY: host_errors
host_errors:
	$(SSH) tail -f "$(H_REM_DIR)/errors.log"

.PHONY: host_db
host_db:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="db") psql -U postgres ${PG_DB}

.PHONY: host_redis
host_redis:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="redis") redis-cli --pass ${REDIS_PASS}

.PHONY: host_service_start
host_service_start:
	$(SSH) sudo systemctl start $(BINARY_SERVICE)

.PHONY: host_service_stop
host_service_stop:
	$(SSH) sudo systemctl stop $(BINARY_SERVICE)

.PHONY: host_metric_cpu
host_metric_cpu:
	hcloud server metrics --type cpu $(APP_HOST)

.PHONY: host_update
host_update:
	git pull
	sed -i -e '/^  lastUpdated:/s/^.*$$/  lastUpdated: $(shell date +%Y-%m-%d)/' $(LANDING_SRC)/config.yaml

.PHONY: host_deploy_op
host_deploy_op: 
	sed -e 's&host-operator&${HOST_OPERATOR}&g; s&work-dir&$(H_REM_DIR)&g; s&etc-dir&$(H_ETC_DIR)&g' $(DEPLOY_HOST_SCRIPTS)/host.service > $(BINARY_SERVICE)
	sed -e 's&binary-name&${BINARY_NAME}&g; s&etc-dir&$(H_ETC_DIR)&g' $(DEPLOY_HOST_SCRIPTS)/start.sh > start.sh
	sudo install -m 400 -o ${HOST_OPERATOR} -g ${HOST_OPERATOR} .env $(H_ETC_DIR)
	sudo install -m 644 $(BINARY_SERVICE) /etc/systemd/system
	sudo install -m 700 -o ${HOST_OPERATOR} -g ${HOST_OPERATOR} $(GO_TARGET) start.sh /usr/local/bin
	rm start.sh $(BINARY_SERVICE) $(GO_TARGET)
	sudo systemctl enable $(BINARY_SERVICE)
	sudo systemctl restart $(BINARY_SERVICE)
	sudo systemctl is-active $(BINARY_SERVICE)

.PHONY: host_update_cert_op
host_update_cert_op:
	sudo certbot certificates
	sudo iptables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo certbot certonly --standalone -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} --agree-tos --no-eff-email
	sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}
	sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} /etc/letsencrypt
	sudo chmod 600 ${CERTS_DIR}/cert.pem ${CERTS_DIR}/privkey.pem
	sudo systemctl restart $(BINARY_SERVICE)
	sudo certbot certificates
	sudo systemctl is-active $(BINARY_SERVICE)

#################################
#            BACKUP             #
#################################

.PHONY: docker_db_redeploy
docker_db_redeploy: docker_stop
	${SUDO} docker rm $(DOCKER_DB_CID) || true
	${SUDO} docker volume remove $(PG_DATA) || true
	${SUDO} docker volume create $(PG_DATA)
	COMPOSE_BAKE=true ${SUDO} docker $(DOCKER_COMPOSE) up -d --build db
	sleep 5

.PHONY: docker_db_backup
docker_db_backup:
	mkdir -p $(DB_BACKUP_DIR)
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_dump --inserts --on-conflict-do-nothing -Fc keycloak \
		> $(DB_BACKUP_DIR)/${PG_DB}_keycloak_$(shell TZ=UTC date +%Y%m%d%H%M%S).dump
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_dump --column-inserts --data-only --on-conflict-do-nothing -n dbtable_schema -Fc ${PG_DB} \
		> $(DB_BACKUP_DIR)/${PG_DB}_app_$(shell TZ=UTC date +%Y%m%d%H%M%S).dump

.PHONY: docker_db_upgrade_op
docker_db_upgrade_op:
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_restore -c -d keycloak \
		< $(DB_BACKUP_DIR)/$(shell ls -Art --ignore "${PG_DB}_app*" $(DB_BACKUP_DIR) | tail -n 1) || true
	${SUDO} docker exec -i $(shell ${SUDO} docker ps -aqf "name=db") psql -U postgres -d ${PG_DB} -c " \
		TRUNCATE TABLE dbtable_schema.users CASCADE; \
		TRUNCATE TABLE dbtable_schema.roles CASCADE; \
		TRUNCATE TABLE dbtable_schema.file_types CASCADE; \
		TRUNCATE TABLE dbtable_schema.budgets CASCADE; \
		TRUNCATE TABLE dbtable_schema.timelines CASCADE; \
		TRUNCATE TABLE dbtable_schema.time_units CASCADE; \
	"
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) pg_restore -a --disable-triggers --superuser=postgres -d ${PG_DB} \
		< $(DB_BACKUP_DIR)/$(shell ls -Art --ignore "${PG_DB}_keycloak*" $(DB_BACKUP_DIR) | tail -n 1) || true

.PHONY: docker_db_upgrade
docker_db_upgrade: docker_db_redeploy docker_db_upgrade_op docker_start

.PHONY: host_db_backup
host_db_backup:
	@mkdir -p "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"
	$(SSH) "cd $(H_REM_DIR) && SUDO=sudo make docker_db_backup"
	rsync ${RSYNC_FLAGS} "$(H_SIGN):$(H_REM_DIR)/$(DB_BACKUP_DIR)/" "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"

.PHONY: host_db_backup_restore
host_db_backup_restore:
	rsync ${RSYNC_FLAGS} "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/" "$(H_SIGN):$(H_REM_DIR)/$(DB_BACKUP_DIR)/"

.PHONY: host_db_upgrade_op
host_db_upgrade_op:
	$(SSH) "cd $(H_REM_DIR) && SUDO=sudo make docker_db_upgrade"

.PHONY: host_db_upgrade
host_db_upgrade: host_service_stop host_db_upgrade_op host_service_start
