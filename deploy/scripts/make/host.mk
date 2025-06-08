#################################
#            HOST INIT          #
#################################

.PHONY: host_up
host_up: 
	@mkdir -p $(HOST_LOCAL_DIR)
	date >> "$(HOST_LOCAL_DIR)/start_time"
	@echo "Tailscale auth key:"; \
	read -s TS_AUTH_KEY; \
	sed -e "s&dummyuser&${HOST_OPERATOR}&g; s&ts-auth-key&$$TS_AUTH_KEY&g; s&host-group&$(H_GROUP)&g; s&project-prefix&${PROJECT_PREFIX}&g; s&project-repo&https://${PROJECT_REPO}.git&g; s&https-port&${GO_HTTPS_PORT}&g; s&http-port&${GO_HTTP_PORT}&g; s&project-dir&$(H_ETC_DIR)&g; s&deploy-scripts&$(DEPLOY_HOST_SCRIPTS)&g; s&binary-name&${BINARY_NAME}&g; s&log-dir&/var/log/${PROJECT_PREFIX}&g; s&node-version&$(NODE_VERSION)&g; s&go-version&$(GO_VERSION)&g;" "$(DEPLOY_HOST_SCRIPTS)/cloud-config.yaml" > "$(HOST_LOCAL_DIR)/cloud-config.yaml"
	cp "$(DEPLOY_HOST_SCRIPTS)/public-firewall.json" "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud firewall create --name "${PROJECT_PREFIX}-public-firewall" --rules-file "$(HOST_LOCAL_DIR)/public-firewall.json" >/dev/null
	hcloud firewall describe "${PROJECT_PREFIX}-public-firewall" -o json > "$(HOST_LOCAL_DIR)/public-firewall.json"
	hcloud server create \
		--name "${APP_HOST}" --datacenter "${HETZNER_DATACENTER}" \
		--type "${HETZNER_APP_TYPE}" --image "${HETZNER_IMAGE}" \
		--user-data-from-file "$(HOST_LOCAL_DIR)/cloud-config.yaml" --firewall "${PROJECT_PREFIX}-public-firewall" >/dev/null
	hcloud server describe "${APP_HOST}" -o json > "$(HOST_LOCAL_DIR)/app.json"
	jq -r '.public_net.ipv6.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip6"
	jq -r '.public_net.ipv4.ip' $(HOST_LOCAL_DIR)/app.json > "$(HOST_LOCAL_DIR)/app_ip"
	@echo "Server is deploying! This will take a moment..."
	@echo "Continue with make host_install_service, once ${APP_HOST} appears in Tailscale and some time is allowed for server reboot."

# ,
#   {
#     "direction": "in",
#     "protocol": "udp",
#     "port": "44400-44500",
#     "source_ips": ["0.0.0.0/0", "::/0"]
#   }

.PHONY: host_install_service
host_install_service: host_sync_files host_update_cert
	$(SSH) "cd $(H_ETC_DIR) && make host_install_service_op"

.PHONY: host_install_service_op
host_install_service_op:
	sudo mkdir -p /etc/logrotate.d/${PROJECT_PREFIX}
	sed -e 's&project-dir&$(H_ETC_DIR)&g;' "$(CRON_SCRIPTS)/whitelist-ips" | sudo tee "/etc/cron.daily/whitelist-ips" >/dev/null
	sudo chmod 755 "/etc/cron.daily/whitelist-ips"
	sed -e 's&project-prefix&${PROJECT_PREFIX}&g;' "$(DEPLOY_HOST_SCRIPTS)/jail.local" | sudo tee /etc/fail2ban/jail.local >/dev/null
	sed -e 's&project-prefix&${PROJECT_PREFIX}&g;' "$(DEPLOY_HOST_SCRIPTS)/logrotate.conf" | sudo tee /etc/logrotate.d/${PROJECT_PREFIX}/logrotate.conf >/dev/null
	sudo cp "$(DEPLOY_HOST_SCRIPTS)/http-auth.conf" "$(DEPLOY_HOST_SCRIPTS)/http-access.conf" /etc/fail2ban/filter.d/
	sudo cp "$(DEPLOY_HOST_SCRIPTS)/ufw-subnet.conf" /etc/fail2ban/action.d/
	sed -e 's&binary-name&${BINARY_NAME}&g; s&etc-dir&$(H_ETC_DIR)&g' "$(DEPLOY_HOST_SCRIPTS)/start.sh" > start.sh
	sed -e 's&host-operator&${HOST_OPERATOR}&g; s&etc-dir&$(H_ETC_DIR)&g' "$(DEPLOY_HOST_SCRIPTS)/host.service" > $(BINARY_SERVICE)
	sudo install -m 750 -o ${HOST_OPERATOR} -g ${HOST_OPERATOR} start.sh /usr/local/bin
	sudo install -m 644 $(BINARY_SERVICE) /etc/systemd/system
	rm start.sh $(BINARY_SERVICE)
	sudo systemctl restart fail2ban
	sudo systemctl enable $(BINARY_SERVICE)

#################################
#           HOST CERTS          #
#################################

# if we don't have certs locally or we're renewing, do the normal cert request and store the certs locally
# if the server still doesn't have certs then we aren't renewing and already have certs locally, likely new deployment
.PHONY: host_update_cert
host_update_cert: host_install_cert host_replace_cert
	@echo "certs handled"

.PHONY: host_install_cert
host_install_cert:
	@if [ ! -d "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}" ] || [ -n "$$RENEW_CERT" ]; then \
		$(SSH) " \
			sudo rm -rf /etc/letsencrypt/archive /etc/letsencrypt/live; \
			sudo iptables -D PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT} || true; \
			sudo certbot certonly --standalone -d ${DOMAIN_NAME} -d www.${DOMAIN_NAME} -m ${ADMIN_EMAIL} --agree-tos --no-eff-email; \
			sudo iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-port ${GO_HTTP_PORT}; \
		"; \
		$(MAKE) host_group_cert; \
		$(MAKE) host_backup_cert; \
		echo "installed certs, renew=$$RENEW_CERT"; \
	else \
		echo "skipping install"; \
	fi

.PHONY: host_replace_cert
host_replace_cert:
	@if $(SSH) "[ ! -d /etc/letsencrypt/archive/${DOMAIN_NAME} ]"; then \
		tailscale file cp "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}/"* "$(APP_HOST):"; \
		$(SSH) " \
			sudo mkdir -p /etc/letsencrypt/archive/${DOMAIN_NAME}/ /etc/letsencrypt/live/${DOMAIN_NAME}/; \
			sudo tailscale file get --conflict=overwrite /etc/letsencrypt/archive/${DOMAIN_NAME}/; \
			sudo find /etc/letsencrypt/archive/${DOMAIN_NAME} -maxdepth 1 -type f | while read file; \
			do sudo ln -s \"\$$file\" /etc/letsencrypt/live/${DOMAIN_NAME}/\$$(basename \"\$$file\"); \
			done \
		"; \
		$(MAKE) host_group_cert; \
		echo "replaced certs into new directories"; \
	else \
		echo "using existing certs"; \
	fi

.PHONY: host_group_cert
host_group_cert:
	$(SSH) " \
		sudo chgrp -R ssl-certs /etc/letsencrypt/live /etc/letsencrypt/archive; \
		sudo chmod -R g+r /etc/letsencrypt/live /etc/letsencrypt/archive; \
		sudo chmod g+x /etc/letsencrypt/live /etc/letsencrypt/archive; \
	"

.PHONY: host_backup_cert
host_backup_cert:
	$(SSH) " \
		mkdir -p \"$(H_ETC_DIR)/$(CERT_BACKUP_DIR)\"; \
		sudo cp -a /etc/letsencrypt/archive/${DOMAIN_NAME}/* \"$(H_ETC_DIR)/$(CERT_BACKUP_DIR)\"; \
		sudo tailscale file cp \"$(H_ETC_DIR)/$(CERT_BACKUP_DIR)/\"* $(shell hostname):; \
	"
	mkdir -p "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}"
	tailscale file get --conflict=overwrite "$(CERT_BACKUP_DIR)/${PROJECT_PREFIX}/"

#################################
#           HOST UTILS          #
#################################

.PHONY: host_sync_files
host_sync_files:
	tailscale file cp .env "$(APP_HOST):"
	$(SSH) " \
		sudo tailscale file get --conflict=overwrite $(H_ETC_DIR)/; \
		sudo chown $(H_LOGIN):$(H_GROUP) $(H_ETC_DIR)/.env; \
	"
	tailscale file cp "$(DEMOS_DIR)/"* "$(APP_HOST):"
	$(SSH) " \
		sudo tailscale file get --conflict=overwrite $(H_ETC_DIR)/$(DEMOS_DIR)/; \
		sudo chown -R $(H_LOGIN):$(H_GROUP) $(H_ETC_DIR)/$(DEMOS_DIR)/; \
	"

.PHONY: host_down
host_down:
	hcloud server delete "${APP_HOST}"
	hcloud firewall delete "${PROJECT_PREFIX}-public-firewall"
	rm -rf $(HOST_LOCAL_DIR)

.PHONY: host_reboot
host_reboot:
	hcloud server reboot "${APP_HOST}"
	until ping -c1 "${APP_HOST}" ; do sleep 5; done
	@echo "rebooted"

.PHONY: host_update
host_update: $(LANDING_SRC)/config.yaml
	git reset --hard HEAD
	git pull
	sed -i -e '/^  lastUpdated:/s/^.*$$/  lastUpdated: $(shell date +%Y-%m-%d)/' $(LANDING_SRC)/config.yaml

.PHONY: host_deploy
host_deploy: go_test_unit host_sync_files 
	$(SSH) "cd $(H_ETC_DIR) && make host_update && make build && make host_service_start_op"

.PHONY: host_service_start
host_service_start:
	$(SSH) "cd $(H_ETC_DIR) && make host_service_start_op"

.PHONY: host_service_start_op
host_service_start_op:
	SUDO=sudo make docker_up
	sudo install -m 700 -o ${HOST_OPERATOR} -g 1000 $(GO_TARGET) /usr/local/bin
	sudo systemctl restart $(BINARY_SERVICE)
	sudo systemctl is-active $(BINARY_SERVICE)

.PHONY: host_service_stop
host_service_stop:
	$(SSH) "cd $(H_ETC_DIR) && make host_service_stop_op"

.PHONY: host_service_stop_op
host_service_stop_op:
	sudo systemctl stop $(BINARY_SERVICE)

.PHONY: host_ssh
host_ssh:
	$(SSH)

.PHONY: host_status
host_status:
	$(SSH) sudo journalctl -n 100 -u ${BINARY_SERVICE} -f

.PHONY: host_errors
host_errors:
	$(SSH) "tail -n 100 -f $(LOG_DIR)/errors.log"

.PHONY: host_db
host_db:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="db") psql -U postgres ${PG_DB}

.PHONY: host_redis
host_redis:
	@$(SSH) sudo docker exec -i $(shell $(SSH) sudo docker ps -aqf name="redis") redis-cli --pass ${REDIS_PASS}

.PHONY: host_metric_cpu
host_metric_cpu:
	hcloud server metrics --type cpu $(APP_HOST)

.PHONY: host_run_cron
host_run_cron:
	$(SSH) 'run-parts /etc/cron.daily'

.PHONY: host_db_backup
host_db_backup:
	@mkdir -p "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"
	$(SSH) 'cd $(H_ETC_DIR) && SUDO=sudo make docker_db_backup && tailscale file cp "$(H_ETC_DIR)/$(DB_BACKUP_DIR)/"* "$(shell hostname)"'
	sudo tailscale file get "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"

.PHONY: host_db_backup_restore
host_db_backup_restore:
	tailscale file cp "$(HOST_LOCAL_DIR)/$(DB_BACKUP_DIR)/"* "$(APP_HOST):"
	$(SSH) "sudo tailscale file get $(H_ETC_DIR)/$(DB_BACKUP_DIR)/"

.PHONY: host_db_restore_op
host_db_restore_op:
	$(SSH) "cd $(H_ETC_DIR) && SUDO=sudo make docker_db_restore"

.PHONY: host_db_restore
host_db_restore: host_service_stop host_db_restore_op host_service_start
