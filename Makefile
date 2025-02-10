ENVFILE?=./.env
include $(ENVFILE)
export $(shell sed 's/=.*//' $(ENVFILE))

# secrets 

ifdef D_PG_PASS
PG_PASS=${D_PG_PASS}
endif

ifdef D_PG_WORKER_PASS
PG_WORKER_PASS=${D_PG_WORKER_PASS}
endif

ifdef D_REDIS_PASS
REDIS_PASS=${D_REDIS_PASS}
endif

ifdef D_KC_PASS
KC_PASS=${D_KC_PASS}
endif

ifdef D_KC_API_CLIENT_SECRET
KC_API_CLIENT_SECRET=${D_KC_API_CLIENT_SECRET}
endif

# go build output

BINARY_OUT=$(BINARY_NAME)
BINARY_TEST=$(BINARY_NAME).test
BINARY_SERVICE=$(BINARY_NAME).service

# source files

JAVA_SRC=./java
TS_SRC=./ts
GO_SRC=./go
LANDING_SRC=./landing

# local directories

CERTS_DIR=certs
JAVA_TARGET_DIR=java/target
JAVA_THEMES_DIR=java/themes
LANDING_BUILD_DIR=landing/public
TS_BUILD_DIR=ts/build
HOST_LOCAL_DIR=deployed/${PROJECT_PREFIX}
GO_GEN_DIR=go/pkg/types
GO_MOCKS_GEN_DIR=go/pkg/mocks
export PLAYWRIGHT_CACHE_DIR=working/playwright

# host locations

H_OP=/home/${HOST_OPERATOR}
H_DOCK=$(H_OP)/bin/docker
H_REM_DIR=$(H_OP)/${PROJECT_PREFIX}
H_ETC_DIR=/etc/${PROJECT_PREFIX}

# build artifacts

PROTO_FILES=$(wildcard proto/*.proto)

TS_API_YAML=ts/openapi.yaml
TS_API_BUILD=ts/src/hooks/api.ts
TS_CONFIG_API=ts/openapi-config.json

CLOUD_CONFIG_OUTPUT=$(HOST_LOCAL_DIR)/cloud-config.yaml

# deploy properties
DOCKER_COMPOSE:=compose -f deploy/scripts/docker/docker-compose.yml --env-file $(ENVFILE)

DEPLOY_SCRIPTS=deploy/scripts
DEV_SCRIPTS=deploy/scripts/dev
DOCKER_SCRIPTS=deploy/scripts/docker
DEPLOY_HOST_SCRIPTS=deploy/scripts/host
AUTH_INSTALL_SCRIPT=deploy/scripts/auth/install.sh
SITE_INSTALLER=deploy/scripts/host/install.sh

# backup related
DB_BACKUP_DIR=backups/db
LATEST_KEYCLOAK_RESTORE := $(DB_BACKUP_DIR)/$(shell ls -Art --ignore "${PG_DB}_app*" $(DB_BACKUP_DIR) | tail -n 1)
LATEST_APP_RESTORE := $(DB_BACKUP_DIR)/$(shell ls -Art --ignore "${PG_DB}_keycloak*" $(DB_BACKUP_DIR) | tail -n 1)

RSYNC_FLAGS=-ave 'ssh -p ${SSH_PORT}'

# APP_IP=$(shell hcloud server ip -6 ${APP_HOST})
APP_IP=$(shell hcloud server ip ${APP_HOST})
APP_IP_B=[$(APP_IP)]

SSH_OP=${HOST_OPERATOR}@$(APP_IP)
# was APP_IP_B
SSH_OP_B=${HOST_OPERATOR}@$(APP_IP)
SSH=ssh -p ${SSH_PORT} -T $(SSH_OP)

LOCAL_UNIX_SOCK_DIR=$(shell pwd)/${UNIX_SOCK_DIR_NAME}
define set_local_unix_sock_dir
	$(eval UNIX_SOCK_DIR=${LOCAL_UNIX_SOCK_DIR})
endef

.PHONY: build clean test_clean test_prep test_gen \
	ts_prep ts ts_test ts_protoc ts_dev \
	go_dev go_test go_test_main go_test_pkg go_coverage \
	docker_up docker_down docker_build docker_start docker_stop \
	docker_db docker_db_start docker_db_backup docker_db_restore docker_cycle docker_db_restore_op \
	docker_redis \
	host_status host_up host_gen host_ssh host_down \
	host_service_stop host_service_start \
	host_deploy host_deploy_env host_deploy_sync host_deploy_docker host_predeploy host_postdeploy host_deploy_compose_up host_deploy_compose_down \
	host_update_cert host_update_cert_op \
	host_db host_db_backup host_db_restore host_db_restore_op \
	host_redis

## Builds

build: clean cert java landing ts go

cert: $(CERTS_DIR)
	chmod +x $(DEV_SCRIPTS)/cert.sh && exec $(DEV_SCRIPTS)/cert.sh

java: $(JAVA_TARGET_DIR)
	mvn -f $(JAVA_SRC) install

# using npm here as pnpm symlinks just hugo and doesn't build correctly
landing: $(LANDING_BUILD_DIR)
	npm --prefix ./landing i
	sed -e 's&project-title&${PROJECT_TITLE}&g; s&app-host-url&${APP_HOST_URL}&g;' "$(LANDING_SRC)/config.yaml.template" > "$(LANDING_SRC)/config.yaml"
	npm run --prefix ./landing build

landing_dev: landing
	chmod +x $(LANDING_SRC)/server.sh && pnpm run --dir $(LANDING_SRC) start

ts_protoc: 
	protoc --proto_path=proto \
		--experimental_allow_proto3_optional \
		--openapi_out=$(TS_SRC) \
		$(PROTO_FILES)
	npx @rtk-query/codegen-openapi $(TS_CONFIG_API)

ts_prep: ts_protoc
	pnpm --dir $(TS_SRC) i
	sed -e 's&app-host-url&${APP_HOST_URL}&g; s&app-host-name&${APP_HOST_NAME}&g; s&kc-realm&${KC_REALM}&g; s&kc-client&${KC_CLIENT}&g; s&kc-path&${KC_PATH}&g; s&turn-name&${TURN_NAME}&g; s&turn-pass&${TURN_PASS}&g; s&allowed-file-ext&${ALLOWED_FILE_EXT}&g;' "$(TS_SRC)/settings.application.env.template" > "$(TS_SRC)/settings.application.env"

ts: ts_prep
	pnpm run --dir $(TS_SRC) build

ts_test: ts_prep
	NODE_ENV=test pnpm run --dir $(TS_SRC) build

ts_dev: ts
	HTTPS=true WDS_SOCKET_PORT=${GO_HTTPS_PORT} pnpm run --dir $(TS_SRC) start

go_protoc: $(GO_GEN_DIR)
	protoc --proto_path=proto \
		--experimental_allow_proto3_optional \
		--go_out=$(GO_SRC) \
		$(PROTO_FILES)

go: go_protoc
	$(call set_local_unix_sock_dir)
	go build -C $(GO_SRC) -o ${PWD}/$(BINARY_OUT) .

go_dev: go cert
	exec ./$(BINARY_NAME) --log debug

go_test: docker_up go_test_main go_test_pkg

go_test_main: $(PLAYWRIGHT_CACHE_DIR) go
	$(call set_local_unix_sock_dir)
	go test -C $(GO_SRC) -v -c -o ../$(BINARY_TEST) && exec ./$(BINARY_TEST)

go_test_pkg: go go_genmocks
	$(call set_local_unix_sock_dir)
	go test -C $(GO_SRC) -v ./...

go_coverage: go_protoc go_genmocks
	go test -C $(GO_SRC) -coverpkg=./... ./...

test_prep: ts_test go_genmocks

test_clean:
	rm -rf $(PLAYWRIGHT_CACHE_DIR)

test_gen:
	npx playwright codegen --ignore-https-errors https://localhost:${GO_HTTPS_PORT}

clean:
	rm -rf $(TS_BUILD_DIR) $(GO_MOCKS_GEN_DIR) $(GO_GEN_DIR) $(JAVA_TARGET_DIR) $(LANDING_BUILD_DIR) $(CERTS_DIR) $(PLAYWRIGHT_CACHE_DIR)
	rm -f $(BINARY_OUT) $(BINARY_TEST) $(TS_API_YAML) $(TS_API_BUILD)

## Tests
go_genmocks: $(GO_MOCKS_GEN_DIR)
	mockgen -source=go/pkg/clients/interfaces.go -destination=$(GO_MOCKS_GEN_DIR)/clients.go -package=mocks

## Utilities

docker_up:
	$(call set_local_unix_sock_dir)
	${SUDO} docker volume create $(PG_DATA)
	${SUDO} docker volume create $(REDIS_DATA)
	${SUDO} docker $(DOCKER_COMPOSE) up -d --build
	chmod +x $(AUTH_INSTALL_SCRIPT) && exec $(AUTH_INSTALL_SCRIPT)

docker_down:
	${SUDO} docker $(DOCKER_COMPOSE) down 
	${SUDO} docker volume remove $(PG_DATA) || true
	${SUDO} docker volume remove $(REDIS_DATA) || true

docker_build:
	$(call set_local_unix_sock_dir)
	${SUDO} docker $(DOCKER_COMPOSE) build

docker_start: docker_build
	${SUDO} docker $(DOCKER_COMPOSE) up -d
	@while true; do if [ $$(curl -o /dev/null -s -w "%{http_code}" "${KC_INTERNAL}") = "303" ]; then break; fi; echo "waiting for keycloak..." && sleep 5; done

docker_stop:
	${SUDO} docker $(DOCKER_COMPOSE) stop 

docker_db:
	${SUDO} docker exec -it $(shell ${SUDO} docker ps -aqf "name=db") psql -U postgres -d ${PG_DB}

docker_db_start:
	${SUDO} docker $(DOCKER_COMPOSE) up -d db
	sleep 5

docker_db_backup: 
	${SUDO} docker exec $(shell ${SUDO} docker ps -aqf "name=db") \
		pg_dump -U postgres -Fc -N dbtable_schema -N dbview_schema -N dbfunc_schema ${PG_DB} > $(DB_BACKUP_DIR)/${PG_DB}_keycloak_$(shell TZ=UTC date +%Y%m%d%H%M%S).dump
	${SUDO} docker exec $(shell ${SUDO} docker ps -aqf "name=db") \
		pg_dump -U postgres --inserts --on-conflict-do-nothing -a -Fc -n dbtable_schema ${PG_DB} > $(DB_BACKUP_DIR)/${PG_DB}_app_$(shell TZ=UTC date +%Y%m%d%H%M%S).dump

docker_db_restore:
	${SUDO} docker exec -i $(shell ${SUDO} docker ps -aqf "name=db") \
		pg_restore -U postgres -c -d ${PG_DB} < $(LATEST_KEYCLOAK_RESTORE)
	${SUDO} docker exec -i $(shell ${SUDO} docker ps -aqf "name=db") \
		pg_restore -U postgres -c -d ${PG_DB} < $(LATEST_APP_RESTORE) 

# ${SUDO} docker exec -i $(shell ${SUDO} docker ps -aqf "name=db") \
# 	psql -U postgres -d ${PG_DB} -c "\
# 		TRUNCATE TABLE dbtable_schema.file_types CASCADE; \
# 		TRUNCATE TABLE dbtable_schema.budgets CASCADE; \
# 		TRUNCATE TABLE dbtable_schema.time_units CASCADE; \
# 		TRUNCATE TABLE dbtable_schema.timelines CASCADE; \
# 	"

docker_cycle: docker_down docker_up

docker_db_restore_op: docker_stop docker_db_start docker_db_restore docker_start

docker_redis:
	${SUDO} docker exec -it $(shell docker ps -aqf "name=redis") redis-cli --pass ${REDIS_PASS}

host_status:
	$(SSH) "sudo journalctl -u ${BINARY_SERVICE} -f"

host_up: host_gen 
	until ping -c1 $(APP_IP) ; do sleep 5; done
	until ssh-keyscan -p ${SSH_PORT} -H $(APP_IP) >> ~/.ssh/known_hosts; do sleep 5; done
	$(SSH) " \
		sudo mkdir -p $(H_ETC_DIR); \
		mkdir -p \
		$(UNIX_SOCK_DIR) \
		$(H_REM_DIR)/$(DB_BACKUP_DIR) \
		$(H_REM_DIR)/$(DEPLOY_SCRIPTS) \
		$(H_REM_DIR)/$(JAVA_TARGET_DIR) \
		$(H_REM_DIR)/$(JAVA_THEMES_DIR) \
		$(H_REM_DIR)/$(TS_BUILD_DIR) \
		$(H_REM_DIR)/$(LANDING_BUILD_DIR); \
	"

# cd $(H_REM_DIR) && mkdir $(UNIX_SOCK_DIR) && sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} $(UNIX_SOCK_DIR); \

host_gen: $(HOST_LOCAL_DIR)
	date >> "$(HOST_LOCAL_DIR)/start_time"
	@sed -e 's&dummyuser&${HOST_OPERATOR}&g; s&id-rsa-pub&$(shell cat ~/.ssh/id_rsa.pub)&g; s&ssh-port&${SSH_PORT}&g; s&https-port&${GO_HTTPS_PORT}&g; s&http-port&${GO_HTTP_PORT}&g;' $(DEPLOY_HOST_SCRIPTS)/cloud-config.yaml > "$(HOST_LOCAL_DIR)/cloud-config.yaml"
	sed -e 's&ssh-port&${SSH_PORT}&g;' $(DEPLOY_HOST_SCRIPTS)/public-firewall.json > "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud firewall create --name "${PROJECT_PREFIX}-public-firewall" --rules-file "$(HOST_LOCAL_DIR)/public-firewall.json" >/dev/null
	hcloud firewall describe "${PROJECT_PREFIX}-public-firewall" -o json > "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud server create \
		--name "${APP_HOST}" --datacenter "${HETZNER_DATACENTER}" \
		--type "${HETZNER_APP_TYPE}" --image "${HETZNER_IMAGE}" \
		--user-data-from-file "$(HOST_LOCAL_DIR)/cloud-config.yaml" --firewall "${PROJECT_PREFIX}-public-firewall" >/dev/null
	hcloud server describe "${APP_HOST}" -o json > "$(HOST_LOCAL_DIR)/app.json"
	jq -r '.public_net.ipv6.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip"

host_down:
	# was APP_IP_B
	ssh-keygen -f ~/.ssh/known_hosts -R "$(APP_IP):${SSH_PORT}"
	hcloud server delete "${APP_HOST}"
	hcloud firewall delete "${PROJECT_PREFIX}-public-firewall"
	rm -rf $(HOST_LOCAL_DIR)

host_ssh:
	ssh -p ${SSH_PORT} ${HOST_OPERATOR}@$(APP_IP)

host_db:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="db") psql -U ${PG_USER} ${PG_DB}

host_redis:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="redis") redis-cli --pass ${REDIS_PASS}

host_cmd:
	$(SSH) $(CMD)

host_service_start:
	$(SSH) sudo systemctl start $(BINARY_SERVICE)

host_service_stop:
	$(SSH) sudo systemctl stop $(BINARY_SERVICE)

host_deploy_env:
	sed -e 's&host-operator&${HOST_OPERATOR}&g; s&work-dir&$(H_REM_DIR)&g; s&etc-dir&$(H_ETC_DIR)&g' $(DEPLOY_HOST_SCRIPTS)/host.service > "$(HOST_LOCAL_DIR)/${BINARY_NAME}.service"
	sed -e 's&binary-name&${BINARY_NAME}&g; s&etc-dir&$(H_ETC_DIR)&g' $(DEPLOY_HOST_SCRIPTS)/start.sh > "$(HOST_LOCAL_DIR)/start.sh"
	rsync ${RSYNC_FLAGS} .env "$(SSH_OP_B):$(H_OP)/.env"
	rsync ${RSYNC_FLAGS} Makefile "$(SSH_OP_B):$(H_REM_DIR)/Makefile"
	rsync ${RSYNC_FLAGS} "$(HOST_LOCAL_DIR)/$(BINARY_SERVICE)" "$(SSH_OP_B):$(H_OP)/$(BINARY_SERVICE)"
	rsync ${RSYNC_FLAGS} "$(HOST_LOCAL_DIR)/start.sh" "$(SSH_OP_B):$(H_OP)/start.sh"
	rsync ${RSYNC_FLAGS} "${BINARY_NAME}" "$(SSH_OP_B):$(H_OP)/$(BINARY_NAME)"
	$(SSH) " \
		if [ ! -f ${CERTS_DIR}/cert.pem ]; then \
			sudo certbot certonly --standalone -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} --agree-tos --no-eff-email; \
			sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} /etc/letsencrypt; \
			sudo chmod 700 ${CERTS_DIR}/cert.pem ${CERTS_DIR}/privkey.pem; \
			sudo ip6tables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
			sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
			sudo ip6tables -A PREROUTING -t nat -p tcp --dport 443 -j REDIRECT --to-port ${GO_HTTPS_PORT}; \
			sudo iptables -A PREROUTING -t nat -p tcp --dport 443 -j REDIRECT --to-port ${GO_HTTPS_PORT}; \
		fi; \
	"
	echo "setting environment properties on the host"
	@$(SSH) " \
		sed -i -e ' \
			/^\(#\|\)PG_PASS=/s&^.*$$&PG_PASS=${PG_PASS}&; \
			/^\(#\|\)PG_WORKER_PASS=/s&^.*$$&PG_WORKER_PASS=${PG_WORKER_PASS}&; \
			/^\(#\|\)REDIS_PASS=/s&^.*$$&REDIS_PASS=${REDIS_PASS}&; \
			/^\(#\|\)KC_PASS=/s&^.*$$&KC_PASS=${KC_PASS}&; \
			/^\(#\|\)KC_API_CLIENT_SECRET=/s&^.*$$&KC_API_CLIENT_SECRET=${KC_API_CLIENT_SECRET}&; \
		' $(H_OP)/.env; \
		echo \"OPENAI_API_KEY=${OPENAI_API_KEY}\" >> $(H_OP)/.env; \
		sudo mv $(H_OP)/.env $(H_ETC_DIR)/.env; \
		sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} $(H_ETC_DIR); \
		sudo chmod -R 700 $(H_ETC_DIR); \
		sudo mv $(H_OP)/$(BINARY_SERVICE) /etc/systemd/system/$(BINARY_SERVICE); \
		sudo chown root:root /etc/systemd/system/$(BINARY_SERVICE); \
		sudo chmod 644 /etc/systemd/system/$(BINARY_SERVICE); \
		sudo mv $(H_OP)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
		sudo chown root:root /usr/local/bin/$(BINARY_NAME); \
		sudo chmod 755 /usr/local/bin/$(BINARY_NAME); \
		sudo mv $(H_OP)/start.sh /usr/local/bin/start.sh; \
		sudo chown root:root /usr/local/bin/start.sh; \
		sudo chmod 755 /usr/local/bin/start.sh; \
	"
	echo "properties set"

host_update_cert:
	$(SSH) " \
		sudo certbot certificates; \
		sudo ip6tables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
		sudo iptables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
		sudo certbot certonly --standalone -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} --agree-tos --no-eff-email; \
		sudo ip6tables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
		sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
		sudo chown -R ${HOST_OPERATOR}:${HOST_OPERATOR} /etc/letsencrypt; \
		sudo chmod 700 ${CERTS_DIR}/cert.pem ${CERTS_DIR}/privkey.pem; \
		sudo systemctl restart $(BINARY_SERVICE); \
		sudo certbot certificates; \
		sudo systemctl is-active $(BINARY_SERVICE); \
	"

host_update_cert_op: host_predeploy host_update_cert host_postdeploy

host_deploy_sync:
	rsync ${RSYNC_FLAGS} Makefile "$(SSH_OP_B):$(H_REM_DIR)/Makefile"
	rsync ${RSYNC_FLAGS} "$(DEPLOY_SCRIPTS)/" "$(SSH_OP_B):$(H_REM_DIR)/$(DEPLOY_SCRIPTS)/"
	rsync ${RSYNC_FLAGS} "$(JAVA_TARGET_DIR)/" "$(SSH_OP_B):$(H_REM_DIR)/$(JAVA_TARGET_DIR)/"
	rsync ${RSYNC_FLAGS} "$(JAVA_THEMES_DIR)/" "$(SSH_OP_B):$(H_REM_DIR)/$(JAVA_THEMES_DIR)/"
	rsync ${RSYNC_FLAGS} "$(TS_BUILD_DIR)/" "$(SSH_OP_B):$(H_REM_DIR)/$(TS_BUILD_DIR)/"
	rsync ${RSYNC_FLAGS} "$(LANDING_BUILD_DIR)/" "$(SSH_OP_B):$(H_REM_DIR)/$(LANDING_BUILD_DIR)/"

host_deploy_docker:
	$(SSH) " \
		if ! command -v docker >/dev/null 2>&1; then \
			curl -fsSL https://get.docker.com | sh; \
		fi \
	"

host_predeploy:
	# was \[APP_IP\[
	sed -i -e "/^\(#\|\)APP_HOST_NAME=/s/^.*$$/APP_HOST_NAME=${DOMAIN_NAME}/;" $(ENVFILE)
	# was APP_IP_B
	$(eval APP_HOST_NAME=${DOMAIN_NAME})
	sed -i -e "/^\(#\|\)CERTS_DIR=/s&^.*$$&CERTS_DIR=/etc/letsencrypt/live/${DOMAIN_NAME}&;" $(ENVFILE)
	$(eval CERTS_DIR=/etc/letsencrypt/live/${DOMAIN_NAME})

host_postdeploy:
	sed -i -e '/^\(#\|\)APP_HOST_NAME=/s/^.*$$/APP_HOST_NAME=localhost:${GO_HTTPS_PORT}/;' $(ENVFILE)
	$(eval APP_HOST_NAME=localhost:${GO_HTTPS_PORT})
	sed -i -e "/^\(#\|\)CERTS_DIR=/s&^.*$$&CERTS_DIR=certs&;" $(ENVFILE)
	$(eval CERTS_DIR=certs)

host_deploy_compose_up:
	$(SSH) "cd $(H_REM_DIR) && SUDO=sudo ENVFILE=$(H_ETC_DIR)/.env make docker_up"

host_deploy_compose_down:
	$(SSH) "cd $(H_REM_DIR) && SUDO=sudo ENVFILE=$(H_ETC_DIR)/.env make docker_down"

host_db_backup:
	$(SSH) "cd $(H_REM_DIR) && SUDO=sudo ENVFILE=$(H_ETC_DIR)/.env make docker_db_backup"

host_db_restore:
	$(SSH) "cd $(H_REM_DIR) && SUDO=sudo ENVFILE=$(H_ETC_DIR)/.env make docker_db_restore_op"

host_db_restore_op: host_predeploy host_service_stop host_db_restore host_service_start host_postdeploy

host_deploy: host_predeploy build host_deploy_env host_deploy_sync host_deploy_docker host_deploy_compose_up host_postdeploy
	$(SSH) " \
		sudo systemctl enable $(BINARY_SERVICE); \
		sudo systemctl stop $(BINARY_SERVICE); \
		sudo systemctl start $(BINARY_SERVICE); \
	"

host_metric_cpu:
	hcloud server metrics --type cpu $(APP_HOST)

$(GO_MOCKS_GEN_DIR) $(GO_GEN_DIR) $(LANDING_BUILD_DIR) $(JAVA_TARGET_DIR) $(HOST_LOCAL_DIR) $(CERTS_DIR) $(DB_BACKUP_DIR) $(PLAYWRIGHT_CACHE_DIR):
	mkdir -p $@

# sed -i -e "/^\(#\|\)UNIX_SOCK_DIR=/s&^.*$$&UNIX_SOCK_DIR=local_tmp&;" $(ENVFILE)
# $(eval UNIX_SOCK_DIR=local_tmp)
# sed -i -e "/^\(#\|\)UNIX_SOCK_DIR=/s&^.*$$&UNIX_SOCK_DIR=\./local_tmp&;" $(ENVFILE)
# $(eval UNIX_SOCK_DIR=./local_tmp)
# sed -i -e "/^\(#\|\)UNIX_SOCK_DIR=/s/^.*$$/UNIX_SOCK_DIR=\/home\/${HOST_OPERATOR}\/local_tmp/;" $(ENVFILE)
# $(eval UNIX_SOCK_DIR=/unix_sock)
# sed -i -e '/^\(#\|\)UNIX_SOCK_DIR=/s/^.*$$/UNIX_SOCK_DIR=\.\/local_tmp/;' $(ENVFILE)
# $(eval UNIX_SOCK_DIR=./local_tmp)
# docker load -i $(AUTH_IMAGE); \
# docker load -i $(DB_IMAGE); \
# docker load -i $(TURN_IMAGE); \
# docker load -i $(REDIS_IMAGE); \
# . $(H_REM_DIR)/$(DOCKER_SCRIPTS)/getdocker.sh && dockerd-rootless-setuptool.sh install; \
# docker compose -f $(H_REM_DIR)/$(DOCKER_SCRIPTS)/docker-compose.yml --env-file $(H_ETC_DIR)/.env up -d; \
#
#		-e KC_SPI_TRUSTSTORE_FILE_HOSTNAME_VERIFICATION_POLICY=ANY \
#		-e KC_HTTPS_CERTIFICATE_FILE=/opt/keycloak/conf/keycloak_fullchain.pem \
#		-e KC_HTTPS_CERTIFICATE_KEY_FILE=/opt/keycloak/conf/keycloak.key \
#		-e KC_SPI_TRUSTSTORE_FILE_FILE=/opt/keycloak/conf/KeyStore.jks \
#		-e KC_SPI_TRUSTSTORE_FILE_PASSWORD=$CA_PASS \
#
# host_deploy: build host_deploy_env host_deploy_sync host_deploy_docker host_deploy_db host_deploy_redis host_deploy_auth
# 	$(SSH) "sudo systemctl enable $(BINARY_SERVICE); sudo systemctl restart $(BINARY_SERVICE);"
#
# host_deploy_db:
# 	$(SSH) " \
# 		sudo docker stop ${DB_IMAGE}_container; \
# 		sudo docker rm ${DB_IMAGE}_container; \
# 		sudo docker volume create ${PG_DATA}; \
# 		sudo docker build -t ${DB_IMAGE} $(H_REM_DIR) -f $(H_REM_DIR)/$(DEPLOY_SCRIPTS)/docker/db_Dockerfile; \
# 		sudo docker run -d --restart always --name ${DB_IMAGE}_container --network host \
# 		-e POSTGRES_DB=${PG_DB} \
# 		-e POSTGRES_USER=${PG_USER} \
# 		-e POSTGRES_PASSWORD=${PG_PASS} \
# 		-v ${PG_DATA}:/var/lib/postgresql/data ${DB_IMAGE}; \
# 	"
# 		
# host_deploy_redis:
# 	$(SSH) " \
# 		sudo docker stop ${REDIS_IMAGE}_container; \
# 		sudo docker rm ${REDIS_IMAGE}_container; \
# 		sudo docker volume create ${REDIS_DATA}; \
# 		sudo docker run -d --restart always --name ${REDIS_IMAGE}_container --network host \
# 		-e REDIS_PASSWORD=${REDIS_PASS} \
# 		-v ${REDIS_DATA}:/data ${REDIS_IMAGE} redis-server --requirepass ${REDIS_PASS}; \
# 	"
#
# host_deploy_auth:
# 	$(SSH) " \
# 		sudo docker stop ${AUTH_IMAGE}_container; \
# 		sudo docker rm ${AUTH_IMAGE}_container; \
# 		sudo docker build -t ${AUTH_IMAGE} $(H_REM_DIR) -f $(H_REM_DIR)/$(DEPLOY_SCRIPTS)/docker/auth_Dockerfile; \
# 		sudo docker run -d --restart always --name ${AUTH_IMAGE}_container --network host \
# 		-e KC_UNIX_SOCK_DIR=$(H_OP)/${UNIX_SOCK_DIR} \
# 		-e KC_UNIX_SOCK_FILE=${UNIX_SOCK_FILE} \
# 		-e KC_API_CLIENT_ID=${KC_API_CLIENT_ID} \
# 		-e KC_PROXY=edge \
# 		-e KC_HTTP_ENABLED=true \
# 		-e KC_HOSTNAME_ADMIN_URL=${KC_PATH} \
# 		-e KC_HOSTNAME_URL=${KC_PATH} \
# 		-e KEYCLOAK_ADMIN=${KC_ADMIN} \
# 		-e KEYCLOAK_ADMIN_PASSWORD=${KC_PASS} \
# 		-e KC_DB_URL=jdbc:postgresql://0.0.0.0:${PG_PORT}/${PG_DB} \
# 		-e KC_DB_USERNAME=${PG_USER} \
# 		-e KC_DB_PASSWORD=${PG_PASS} \
# 		-e KC_REDIS_HOST=0.0.0.0 \
# 		-e KC_REDIS_PORT=${REDIS_PORT} \
# 		-e KC_REDIS_PASS=${REDIS_PASS} \
# 		-e KC_REGISTRATION_RATE_LIMIT=10 \
# 		-v $(H_OP)/${UNIX_SOCK_DIR}:/${UNIX_SOCK_DIR} ${AUTH_IMAGE}; \
# 		SUDO=sudo APP_HOST_URL=$(APP_HOST_URL) . $(H_REM_DIR)/$(AUTH_INSTALL_SCRIPT); \
# 	"
#
# host_deploy_turn:
# 	$(SSH) ""
