LLM_RO_PATHS:=.env.template .env .git .gitignore README.md log go/pkg/types $(foreach a,llm db $(foreach b,vars functions test deps,make/$(b).mk),deploy/scripts/$(a))
LLM_RW_PATHS:=go java landing proto ts
LLM_NO_PATHS:=ts/node_modules landing/node_modules proto/validate proto/google go/buf.build

LLM_BASE_VOLUMES_RW = $(subst $(space),$(empty),$(strip $(patsubst %,%,$(foreach path,$(LLM_RW_PATHS),$(PROJECT_DIR)/$(path):/workspace/$(path):rw,))))
LLM_BASE_VOLUMES_RO = $(subst $(space),$(empty),$(strip $(patsubst %,%,$(foreach path,$(LLM_RO_PATHS),$(PROJECT_DIR)/$(path):/workspace/$(path):ro,))))

LLM_BASE_VOLUMES_RW := $(patsubst %$(comma),%,$(LLM_BASE_VOLUMES_RW))
LLM_BASE_VOLUMES_RO := $(patsubst %$(comma),%,$(LLM_BASE_VOLUMES_RO))

LLM_NO_VOLUMES=$(foreach path,$(LLM_NO_PATHS),--tmpfs /workspace/$(path))

LLM_VOLUMES=-e SANDBOX_VOLUMES=$(PROJECT_DIR)/deploy/scripts/llm/Makefile:/workspace/Makefile,$(LLM_BASE_VOLUMES_RW),$(LLM_BASE_VOLUMES_RO) $(LLM_NO_VOLUMES)

#################################
#              LLM              #
#################################

.PHONY: llm_review
llm_review:
	cat working/test_thing > working/llm_text
	@while IFS= read -r line || [ -n "$$line" ]; do \
		echo -e "$$line" >> working/llm_text; \
	done

.PHONY: llm_fix
llm_fix:
	echo "$(shell cat working/llm_text)" | $(MAKE) llm_ask

.PHONY: llm_clean
llm_clean:
	docker stop $(shell docker ps -aqf "name=openhands") && docker rm $(shell docker ps -aqf "name=openhands") || true

.PHONY: llm_ask
llm_ask: llm_clean
	@docker run -t --rm --pull=always \
	$(LLM_VOLUMES) \
	--network host \
	-e SANDBOX_USER_ID=$$(id -u) \
	-e SANDBOX_RUNTIME_CONTAINER_IMAGE=docker.all-hands.dev/all-hands-ai/runtime:0.41-nikolaik \
	-e LLM_API_KEY=$(GEMINI_2_5_KEY) \
	-e LLM_MODEL="gemini/gemini-2.0-flash" \
	-e LOG_ALL_EVENTS=true \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v ~/.openhands-state:/.openhands-state \
	--add-host host.docker.internal:host-gateway \
	--name openhands-app \
	docker.all-hands.dev/all-hands-ai/openhands:0.41 \
	python -m openhands.core.main -t "$(shell cat) $(shell cat $(LLM_SCRIPTS)/INSTRUCTIONS.md)"

# -e SANDBOX_USER_ID=$(shell id -u) \

.PHONY: llm_ui
llm_ui:
	@docker run -it --rm --pull=always \
	-e SANDBOX_USE_HOST_NETWORK=true \
	-e SANDBOX_USER_ID=$$(id -u) \
	-e SANDBOX_RUNTIME_CONTAINER_IMAGE=docker.all-hands.dev/all-hands-ai/runtime:0.41-nikolaik \
	-e LOG_ALL_EVENTS=true \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v ~/.openhands-state:/.openhands-state \
	-p 3000:3000 \
	--add-host host.docker.internal:host-gateway \
	--name openhands-app \
	docker.all-hands.dev/all-hands-ai/openhands:0.41
