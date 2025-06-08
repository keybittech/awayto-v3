define clean_logs
  $(shell if [ $$(ls -1 $(LOG_DIR)/db/*.log 2>/dev/null | wc -l) -gt 1 ]; then \
    ls -t $(LOG_DIR)/db/*.log | tail -n +2 | xargs rm -f; \
  fi)
	rm -f log/*.log
endef

# $(DOCKER_REDIS_EXEC) FLUSHALL
define clean_test
	@echo "======== cleaning tests ========"
	$(call clean_logs)
	$(call set_local_unix_sock_dir)
endef

define set_local_unix_sock_dir
	$(eval UNIX_SOCK_DIR := $(shell pwd)/$(ORIGINAL_SOCK_DIR))
	$(eval TARGET_GROUP := $(if $(filter true,$(DEPLOYING)),$(H_GROUP),1000))
	setfacl -m g:$(TARGET_GROUP):rwx "$(UNIX_SOCK_DIR)"
	setfacl -d -m g:$(TARGET_GROUP):rwx "$(UNIX_SOCK_DIR)"
endef
