SHELL := /bin/bash
ENVFILE?=./.env

$(shell if [ ! -f $(ENVFILE) ]; then install -m 600 .env.template $(ENVFILE); fi)

include $(ENVFILE)
export $(shell sed 's/=.*//' $(ENVFILE))

GO_VERSION=go1.24.3.linux-amd64
NODE_VERSION=v22.13.1

export PATH := ${PATH}:/home/$(shell whoami)/.nvm/versions/node/$(NODE_VERSION)/bin:/home/$(shell whoami)/go/bin:/usr/local/go/bin

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
LOG_DIR=log
HOST_LOCAL_DIR=sites/${PROJECT_PREFIX}
GO_GEN_DIR=$(GO_SRC)/pkg/types
GO_CMD_DIR=$(GO_SRC)/cmd/generate
GO_PROTO_MUTEX_CMD_DIR=$(GO_CMD_DIR)/proto_mutex
GO_HANDLERS_REGISTER_CMD_DIR=$(GO_CMD_DIR)/handlers_register
GO_API_DIR=$(GO_SRC)/pkg/api
GO_CLIENTS_DIR=$(GO_SRC)/pkg/clients
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


BINARY_TEST=$(BINARY_NAME).test
BINARY_SERVICE=$(BINARY_NAME).service
JAVA_TARGET=$(JAVA_TARGET_DIR)/kc-custom.jar
LANDING_TARGET=$(LANDING_BUILD_DIR)/index.html
TS_TARGET=$(TS_BUILD_DIR)/index.html

GO_FILE_DIRS=$(GO_SRC) $(GO_HANDLERS_REGISTER_CMD_DIR) $(GO_GEN_DIR) $(GO_API_DIR) $(GO_CLIENTS_DIR) $(GO_HANDLERS_DIR) $(GO_UTIL_DIR)
GO_FILES=$(foreach dir,$(GO_FILE_DIRS),$(wildcard $(dir)/*.go))
GO_TARGET=${PWD}/$(GO_SRC)/$(BINARY_NAME)
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
DOCKER_REDIS_CID=$(shell ${SUDO} docker ps -aqf "name=redis")
DOCKER_REDIS_EXEC=${SUDO} docker exec -i $(DOCKER_REDIS_CID) redis-cli --pass $$(cat $$REDIS_PASS_FILE)
DOCKER_DB_CID=$(shell ${SUDO} docker ps -aqf "name=db")
DOCKER_DB_EXEC=${SUDO} docker exec --user postgres -it
DOCKER_DB_CMD=${SUDO} docker exec --user postgres -i

#################################
#          DEPLOY PROPS         #
#################################

DEPLOY_SCRIPTS=deploy/scripts
DB_SCRIPTS=$(DEPLOY_SCRIPTS)/db
DEV_SCRIPTS=$(DEPLOY_SCRIPTS)/dev
DOCKER_SCRIPTS=$(DEPLOY_SCRIPTS)/docker
DEPLOY_HOST_SCRIPTS=$(DEPLOY_SCRIPTS)/host
CRON_SCRIPTS=$(DEPLOY_SCRIPTS)/cron
LLM_SCRIPTS=$(DEPLOY_SCRIPTS)/llm
DOCKER_COMPOSE_SCRIPT=$(DEPLOY_SCRIPTS)/docker/docker-compose.yml
AUTH_SCRIPTS=$(DEPLOY_SCRIPTS)/auth
AUTH_INSTALL_SCRIPT=$(AUTH_SCRIPTS)/install.sh

H_ETC_DIR=/etc/${PROJECT_PREFIX}

H_LOGIN=${HOST_OPERATOR}login
H_GROUP=${PROJECT_PREFIX}g
H_SIGN=$(H_LOGIN)@${APP_HOST}
SSH=tailscale ssh $(H_SIGN)

CURRENT_USER:=$(shell whoami)
DEPLOYING:=$(if $(filter ${HOST_OPERATOR}login,$(CURRENT_USER)),true,)

define if_deploying
$(if $(DEPLOYING),$(1),$(2))
endef

ifeq ($(ORIGINAL_SOCK_DIR),)
ORIGINAL_SOCK_DIR:=${UNIX_SOCK_DIR}
endif

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

