#################################
#              LLM              #
#################################

SANDBOX_IMAGE_NAME=awayto-sandbox
SANDBOX_DOCKERFILE=$(LLM_SCRIPTS)/sandbox_Dockerfile

LLM_RO_PATHS:=.env.template .gitignore README.md Makefile deploy secrets
LLM_RW_PATHS:=go java landing proto ts .git log

LLM_BASE_VOLUMES_RW = $(subst $(space),$(empty),$(strip $(patsubst %,%,$(foreach path,$(LLM_RW_PATHS),$(PROJECT_DIR)/$(path):/workspace/$(path):rw,))))
LLM_BASE_VOLUMES_RO = $(subst $(space),$(empty),$(strip $(patsubst %,%,$(foreach path,$(LLM_RO_PATHS),$(PROJECT_DIR)/$(path):/workspace/$(path):ro,))))

LLM_BASE_VOLUMES_RW := $(patsubst %$(comma),%,$(LLM_BASE_VOLUMES_RW))
LLM_BASE_VOLUMES_RO := $(patsubst %$(comma),%,$(LLM_BASE_VOLUMES_RO))

DOCKER_SOCK=/var/run/docker.sock:/var/run/docker.sock:rw
OPENHANDS_CONFIG=$(PROJECT_DIR)/deploy/scripts/llm/.openhands:/workspace/.openhands:ro
LLM_VOLUMES=-e SANDBOX_VOLUMES=$(OPENHANDS_CONFIG),$(DOCKER_SOCK),$(LLM_BASE_VOLUMES_RW),$(LLM_BASE_VOLUMES_RO)

.PHONY: llm_build_sandbox
llm_build_sandbox:
	@echo "Building custom sandbox image..."
	docker build --build-arg GO_VERSION=$(GO_VERSION) -f $(SANDBOX_DOCKERFILE) -t $(SANDBOX_IMAGE_NAME) .
	@echo "Custom sandbox image built: $(SANDBOX_IMAGE_NAME)"

.PHONY: llm_clean
llm_clean:
	docker stop $(shell docker ps -aqf "name=openhands") && docker rm $(shell docker ps -aqf "name=openhands") || true

.PHONY: llm_ui
llm_ui: llm_build_sandbox
	@docker run -it --rm --pull=always \
	$(LLM_VOLUMES) \
	-e SANDBOX_RUNTIME_CONTAINER_IMAGE=$(SANDBOX_IMAGE_NAME) \
	-e LOG_ALL_EVENTS=true \
	-e LOG_LEVEL=DEBUG \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v ~/.openhands-state:/.openhands-state \
	-p 3000:3000 \
	--add-host host.docker.internal:host-gateway \
	--name openhands-app \
	docker.all-hands.dev/all-hands-ai/openhands:0.42

.PHONY: llm_rebuild
llm_rebuild: llm_clean
	docker rmi $(SANDBOX_IMAGE_NAME) || true
	$(MAKE) llm_ui



# -e SANDBOX_RUNTIME_BINDING_ADDRESS=127.0.0.1 \
# -e SANDBOX_USER_ID=$$(id -u) \

.PHONY: aider
aider:
	aider --model gemini/gemini-2.5-flash --add-gitignore-files --no-auto-commits --no-dirty-commits --no-analytics --analytics-disable --watch-files
