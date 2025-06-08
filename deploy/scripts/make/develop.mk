#################################
#           DEVELOP             #
#################################

.PHONY: go_dev
go_dev: 
	$(call clean_logs)
	$(call set_local_unix_sock_dir)
	$(GO_DEV_FLAGS) gow -e=go,mod run -C $(GO_SRC) . $(NO_LIMIT) $(LOG_DEBUG)

.PHONY: go_dev_ts
go_dev_ts: 
	$(call clean_logs)
	$(call set_local_unix_sock_dir)
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
	$(call set_local_unix_sock_dir)
	$(GO_DEV_FLAGS) gow -e=go,mod build -C $(GO_SRC) -o $(GO_TARGET) .

.PHONY: go_debug
go_debug:
	$(call clean_logs)
	$(call set_local_unix_sock_dir)
	cd go && $(GO_DEV_FLAGS) dlv debug --wd ${PWD}

.PHONY: go_debug_exec
go_debug_exec:
	$(call clean_logs)
	$(call set_local_unix_sock_dir)
	$(GO_DEV_FLAGS) dlv exec --wd ${PWD} $(GO_TARGET)

.PHONY: ts_dev
ts_dev:
	BROWSER=none HTTPS=true WDS_SOCKET_PORT=${GO_HTTPS_PORT} pnpm run --dir $(TS_SRC) start

.PHONY: landing_dev
landing_dev: build
	chmod +x $(LANDING_SRC)/server.sh && pnpm run --dir $(LANDING_SRC) start

