CREATE UNIQUE INDEX unique_enabled_name_created_sub ON dbtable_schema.schedules (name, created_sub) WHERE (enabled = true);

CREATE UNIQUE INDEX unique_group_owner ON dbtable_schema.groups (created_sub) WHERE (created_sub IS NOT NULL);
CREATE UNIQUE INDEX unique_code ON dbtable_schema.groups (lower(code));
