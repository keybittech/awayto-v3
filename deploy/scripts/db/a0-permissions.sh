#!/bin/bash

psql -v ON_ERROR_STOP=1 <<-EOSQL

  DROP DATABASE IF EXISTS $PG_DB;
  DROP DATABASE IF EXISTS keycloak;
  DROP USER IF EXISTS $PG_WORKER;

  CREATE USER $PG_WORKER WITH PASSWORD '$(cat $PG_WORKER_PASS_FILE)';
  ALTER ROLE $PG_WORKER VALID UNTIL 'infinity';

  CREATE DATABASE keycloak;

  GRANT ALL ON DATABASE keycloak TO $PG_WORKER;
  ALTER DATABASE keycloak OWNER TO $PG_WORKER;

  CREATE DATABASE $PG_DB;

  ALTER DATABASE $PG_DB SET intervalstyle = 'iso_8601';

  GRANT ALL ON DATABASE $PG_DB TO $PG_WORKER;

EOSQL

psql -v ON_ERROR_STOP=1 --dbname $PG_DB <<-EOSQL
  CREATE SCHEMA dbfunc_schema;
  CREATE SCHEMA dbtable_schema;
  CREATE SCHEMA dbview_schema;

  GRANT USAGE ON SCHEMA dbfunc_schema TO $PG_WORKER;
  GRANT USAGE ON SCHEMA dbtable_schema TO $PG_WORKER;
  GRANT USAGE ON SCHEMA dbview_schema TO $PG_WORKER;

  ALTER DEFAULT PRIVILEGES IN SCHEMA dbfunc_schema GRANT EXECUTE ON FUNCTIONS TO $PG_WORKER;
  ALTER DEFAULT PRIVILEGES IN SCHEMA dbtable_schema GRANT ALL ON TABLES TO $PG_WORKER;
  ALTER DEFAULT PRIVILEGES IN SCHEMA dbview_schema GRANT ALL ON TABLES TO $PG_WORKER;
EOSQL
