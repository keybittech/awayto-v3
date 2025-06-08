#################################
#            DOCKER             #
#################################

.PHONY: docker_up
docker_up: build
	$(call set_local_unix_sock_dir)
	${SUDO} docker volume create $(PG_DATA) || true
	${SUDO} docker volume create $(REDIS_DATA) || true
	COMPOSE_BAKE=true ${SUDO} docker $(DOCKER_COMPOSE) up -d --build
	chmod +x $(AUTH_INSTALL_SCRIPT) && exec $(AUTH_INSTALL_SCRIPT)
	
.PHONY: docker_down
docker_down:
	$(call set_local_unix_sock_dir)
	${SUDO} docker $(DOCKER_COMPOSE) down -v
	${SUDO} docker volume remove $(PG_DATA) || true
	${SUDO} docker volume remove $(REDIS_DATA) || true

.PHONY: docker_cycle
docker_cycle: docker_down docker_up docker_db_backup

.PHONY: docker_build
docker_build:
	$(call set_local_unix_sock_dir)
	${SUDO} docker $(DOCKER_COMPOSE) build

.PHONY: docker_start
docker_start: docker_build
	$(call set_local_unix_sock_dir)
	${SUDO} docker $(DOCKER_COMPOSE) up -d

.PHONY: docker_stop
docker_stop:
	$(call set_local_unix_sock_dir)
	${SUDO} docker $(DOCKER_COMPOSE) stop 

.PHONY: docker_db
docker_db:
	$(DOCKER_DB_EXEC) $(DOCKER_DB_CID) psql -U postgres -d ${PG_DB}

.PHONY: docker_redis
docker_redis:
	${SUDO} docker exec -it $(shell docker ps -aqf "name=redis") redis-cli --pass $$(cat $$REDIS_PASS_FILE)

#################################
#         DOCKER BACKUP         #
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

.PHONY: docker_db_restore_op
docker_db_restore_op:
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

.PHONY: docker_db_restore
docker_db_restore: docker_db_redeploy docker_db_restore_op docker_start

.PHONY: docker_db_redeploy_views
docker_db_redeploy_views:
	@> working/sql_views
	@printf "DROP SCHEMA dbview_schema CASCADE;\nCREATE SCHEMA dbview_schema;\nGRANT USAGE ON SCHEMA dbview_schema TO $(PG_WORKER);\nALTER DEFAULT PRIVILEGES IN SCHEMA dbview_schema GRANT ALL ON TABLES TO $(PG_WORKER);\n\n" >> working/sql_views
	@cat $(DB_SCRIPTS)/base_views.sql >> working/sql_views
	@echo "" >> working/sql_views
	@cat $(DB_SCRIPTS)/app_views.sql >> working/sql_views
	@echo "" >> working/sql_views
	@cat $(DB_SCRIPTS)/function_views.sql >> working/sql_views
	@echo "" >> working/sql_views
	@cat $(DB_SCRIPTS)/kiosk_views.sql >> working/sql_views
	$(DOCKER_DB_CMD) $(DOCKER_DB_CID) psql -U postgres -d ${PG_DB} < working/sql_views

.PHONY: check_logs
check_logs:
	exec $(CRON_SCRIPTS)/check-logs
