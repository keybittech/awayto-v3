#!/bin/bash

psql -v ON_ERROR_STOP=1 <<-EOSQL
  \c $PG_DB $PG_WORKER;

  CREATE UNIQUE INDEX unique_enabled_name_created_sub ON dbtable_schema.schedules (name, created_sub) WHERE (enabled = true);

  CREATE UNIQUE INDEX unique_group_owner ON dbtable_schema.groups (created_sub) WHERE (created_sub IS NOT NULL);
  CREATE UNIQUE INDEX unique_code ON dbtable_schema.groups (lower(code));

  ALTER DATABASE $PG_DB SET intervalstyle = 'iso_8601';
EOSQL
