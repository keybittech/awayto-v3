LLM_BASE_VOLUMES=$(PROJECT_DIR)/.env.template:/workspace/.env.template:rw,$(PROJECT_DIR)/.env:/workspace/.env:ro,$(PROJECT_DIR)/.gitignore:/workspace/.gitignore:ro,$(PROJECT_DIR)/Makefile:/workspace/Makefile:ro,$(PROJECT_DIR)/README.md:/workspace/README.md:ro,$(PROJECT_DIR)/secrets:/workspace/secrets:ro,$(PROJECT_DIR)/deploy:/workspace/deploy:rw,$(PROJECT_DIR)/certs:/workspace/certs:ro,$(PROJECT_DIR)/log:/workspace/log:rw,$(PROJECT_DIR)/demos:/workspace/demos:ro,$(PROJECT_DIR)/local_tmp:/workspace/local_tmp:rw,$(PROJECT_DIR)/go:/workspace/go:rw,$(PROJECT_DIR)/java:/workspace/java:rw,$(PROJECT_DIR)/landing:/workspace/landing:rw,$(PROJECT_DIR)/proto:/workspace/proto:rw,$(PROJECT_DIR)/ts:/workspace/ts:rw

LLM_VOLUMES=-e SANDBOX_VOLUMES=$(LLM_BASE_VOLUMES)

#################################
#              LLM              #
#################################

.PHONY: llm_install_deps
llm_install_deps:
	@cd /workspace
	@echo "Updating package lists..."
	@sudo apt-get update
	@echo "Upgrading existing packages..."
	@sudo apt-get upgrade -y
	@sudo apt-get install -y \
			uidmap \
			build-essential \
			jq \
			default-jre \
			maven \
			hugo \
			protobuf-compiler \
			protoc-gen-go \
			git \
			curl \
			acl
	@echo "Installing Node.js via NVM..."
	curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
	@export NVM_DIR="$$HOME/.nvm"
	[ -s "$$NVM_DIR/nvm.sh" ] && \. "$$NVM_DIR/nvm.sh"
	[ -s "$$NVM_DIR/bash_completion" ] && \. "$$NVM_DIR/bash_completion"
	nvm install $(NODE_VERSION)
	npm i -g pnpm@latest-10
	@echo "Installing Go..."
	@rm -rf /usr/local/go
	curl -L -o /tmp/goinstall.tar.gz https://go.dev/dl/$(GO_VERSION).tar.gz  # Replace with actual go-version
	@tar -C /usr/local -xzf /tmp/goinstall.tar.gz
	@rm /tmp/goinstall.tar.gz
	@echo "Installing Go tools..."
	/usr/local/go/bin/go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	@echo "Installing Docker..."
	curl -fsSL https://get.docker.com | sh
	@echo "Setting up rootless Docker..."
	dockerd-rootless-setuptool.sh install || true

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
	-e SANDBOX_RUNTIME_CONTAINER_IMAGE=docker.all-hands.dev/all-hands-ai/runtime:0.39-nikolaik \
	-e SANDBOX_USER_ID=$(shell id -u) \
	-e LLM_API_KEY=$(GEMINI_2_5_KEY) \
	-e LLM_MODEL="gemini/gemini-2.0-flash" \
	-e LOG_ALL_EVENTS=true \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v ~/.openhands-state:/.openhands-state \
	--add-host host.docker.internal:host-gateway \
	--name openhands-app \
	docker.all-hands.dev/all-hands-ai/openhands:0.39 \
	python -m openhands.core.main -t "$(shell cat) $(shell cat $(LLM_SCRIPTS)/INSTRUCTIONS.md)"
