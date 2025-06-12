.PHONY: install_deps
install_deps:
	echo "Updating package lists..." && \
	sudo apt-get update && \
	echo "Upgrading existing packages..." && \
	sudo apt-get upgrade -y && \
	sudo apt-get install -y \
		build-essential \
		jq \
		acl && \
	echo "Installing Go..." && \
	sudo rm -rf /usr/local/go && \
	curl -L -o /tmp/goinstall.tar.gz https://go.dev/dl/$(GO_VERSION).tar.gz && \
	sudo tar -C /usr/local -xzf /tmp/goinstall.tar.gz && \
	rm /tmp/goinstall.tar.gz && \
	echo "Installing Go tools..." && \
	/usr/local/go/bin/go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest || true


.PHONY: install_deps_old
install_deps_old:
	echo "Updating package lists..." && \
	sudo apt-get update && \
	echo "Upgrading existing packages..." && \
	sudo apt-get upgrade -y && \
	sudo apt-get install -y \
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
		acl && \
	echo "Installing Node.js via NVM..." && \
	curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash && \
	export NVM_DIR="$$HOME/.nvm" && \
	[ -s "$$NVM_DIR/nvm.sh" ] && . "$$NVM_DIR/nvm.sh" && \
	[ -s "$$NVM_DIR/bash_completion" ] && . "$$NVM_DIR/bash_completion" && \
	nvm install $(NODE_VERSION) && \
	npm i -g pnpm@latest-10 && \
	echo "Installing Go..." && \
	sudo rm -rf /usr/local/go && \
	curl -L -o /tmp/goinstall.tar.gz https://go.dev/dl/$(GO_VERSION).tar.gz && \
	sudo tar -C /usr/local -xzf /tmp/goinstall.tar.gz && \
	rm /tmp/goinstall.tar.gz && \
	echo "Installing Go tools..." && \
	/usr/local/go/bin/go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest && \
	echo "Installing Docker..." && \
	curl -fsSL https://get.docker.com | sh && \
	echo "Setting up rootless Docker..." && \
	dockerd-rootless-setuptool.sh install || true
