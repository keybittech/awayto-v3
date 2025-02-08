#!/bin/bash

psql -v ON_ERROR_STOP=1 <<-EOSQL
  \c $PG_DB $PG_WORKER;

	CREATE TRIGGER set_group_code AFTER INSERT ON dbtable_schema.groups FOR EACH ROW EXECUTE FUNCTION dbfunc_schema.make_group_code();

EOSQL
