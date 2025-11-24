#!/bin/sh
set -e

# if [ -d "/var/run/postgresql" ]; then
#   chown postgres:root /var/run/postgresql
#   chmod 2770 /var/run/postgresql
# fi
#
# if [ -f "/etc/secrets/pg_worker_pass" ]; then
#   chgrp postgres /etc/secrets/pg_worker_pass
#   chmod 640 /etc/secrets/pg_worker_pass
# fi

if [ -d "/var/run/postgresql" ]; then
  # 1. Owner 'postgres' (100069) so DB can write.
  # 2. Group 'root' (1000/You) so YOU can enter.
  chown postgres:root /var/run/postgresql
  
  # 3. Mode 770: Only Owner and Group can enter. 
  # 'Others' on the host are blocked at the door.
  chmod 770 /var/run/postgresql
fi

if [ -d "/tmp/init_sql" ]; then
  chown -R postgres:postgres /tmp/init_sql
fi

exec /usr/local/bin/docker-entrypoint.sh "$@"

  
