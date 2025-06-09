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
go_test_unit: go_test_unit_api go_test_unit_clients go_test_unit_handlers go_test_unit_util
	@echo "unit tests complete"

.PHONY: go_test_fuzz
go_test_fuzz:
	@mkdir -p $(GO_FUZZ_CACHEDIR)
	@$(call clean_test)
	$(GO) test -C $(GO_HANDLERS_DIR) -c -o handlers.fuzz.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_HANDLERS_DIR)/handlers.fuzz.$(BINARY_TEST) $(GO_FUZZ_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_ui
go_test_ui: docker_db_restore_op
	rm -f demos/*.webm
	$(call clean_test)
	$(GO) test -C $(GO_PLAYWRIGHT_DIR) -c -o playwright.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_PLAYWRIGHT_DIR)/playwright.$(BINARY_TEST) $(NO_LIMIT) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_integration
go_test_integration:
	$(call clean_test)
	$(GO) test -C $(GO_INTEGRATIONS_DIR) -short -c -o integration.$(BINARY_TEST)
	-$(GO_ENVFILE_FLAG) exec $(GO_INTEGRATIONS_DIR)/integration.$(BINARY_TEST) $(GO_TEST_EXEC_FLAGS)
	@cat $(LOG_DIR)/errors.log

.PHONY: go_test_integration_long
go_test_integration_long:
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
go_test_bench: go_test_bench_build
	$(call clean_test)
	-$(GO_ENVFILE_FLAG) exec $(GO_API_DIR)/api.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)
	-$(GO_ENVFILE_FLAG) exec $(GO_CLIENTS_DIR)/clients.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)
	-$(GO_ENVFILE_FLAG) exec $(GO_HANDLERS_DIR)/handlers.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)
	-$(GO_ENVFILE_FLAG) exec $(GO_UTIL_DIR)/util.bench.$(BINARY_TEST) $(GO_BENCH_EXEC_FLAGS)

