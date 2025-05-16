#!/bin/bash

PG_WORKER_PASS="$(cat "$PG_WORKER_PASS_FILE")"

ENV_LIST="PG_DB PG_WORKER_PASS PG_WORKER USER_SUB GROUP_ID IS_WORKER IS_USER IS_CREATOR HAS_GROUP IS_GROUP_ADMIN IS_GROUP_BOOKINGS IS_GROUP_SCHEDULES IS_GROUP_SERVICES IS_GROUP_SCHEDULE_KEYS IS_GROUP_ROLES IS_GROUP_USERS IS_GROUP_PERMISSIONS"

for f in "$SCRIPT_DIR"/*.sql; do
  for var in $ENV_LIST; do
    val="${!var}"
    sed -i "s|\$$var|$val|g" "$f"
  done
done

psql -v ON_ERROR_STOP=1 \
  -f $SCRIPT_DIR/create_dbs.sql

psql -v ON_ERROR_STOP=1 \
  --dbname $PG_DB \
  -f $SCRIPT_DIR/create_schema_permissions.sql \
  -f $SCRIPT_DIR/base_tables.sql \
  -f $SCRIPT_DIR/base_views.sql \
  -f $SCRIPT_DIR/app_tables.sql \
  -f $SCRIPT_DIR/app_views.sql \
  -f $SCRIPT_DIR/functions.sql \
  -f $SCRIPT_DIR/function_views.sql \
  -f $SCRIPT_DIR/kiosk_views.sql \
  -f $SCRIPT_DIR/app_alterations.sql \
  -f $SCRIPT_DIR/triggers.sql

rm -rf $SCRIPT_DIR

exit 0
