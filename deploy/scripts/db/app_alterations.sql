CREATE UNIQUE INDEX unique_enabled_name_created_sub ON dbtable_schema.schedules (name, created_sub) WHERE (enabled = true);

CREATE UNIQUE INDEX unique_group_owner ON dbtable_schema.groups (created_sub) WHERE (created_sub IS NOT NULL);
CREATE UNIQUE INDEX unique_code ON dbtable_schema.groups (lower(code));

-- use security invoker for all views
DO $$
DECLARE
  r RECORD;
BEGIN
  -- Loop through all views in your specific schema(s)
  FOR r IN
    SELECT schemaname, viewname
    FROM pg_catalog.pg_views
    WHERE schemaname IN ('dbview_schema') -- Add other schemas here if necessary
  LOOP
    -- Dynamically alter the view
    EXECUTE format('ALTER VIEW %I.%I SET (security_invoker = true);', r.schemaname, r.viewname);
  END LOOP;
END;
$$ LANGUAGE plpgsql;
