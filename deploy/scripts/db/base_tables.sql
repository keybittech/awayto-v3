-- from https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
CREATE OR REPLACE FUNCTION dbfunc_schema.uuid_generate_v7() RETURNS uuid
AS $$
BEGIN
  -- use random v4 uuid as starting point (which has the same variant we need)
  -- then overlay timestamp
  -- then set version 7 by flipping the 2 and 1 bit in the version 4 string
  return encode(
    set_bit(
      set_bit(
        overlay(uuid_send(gen_random_uuid())
                placing substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
                from 1 for 6
        ),
        52, 1
      ),
      53, 1
    ),
    'hex')::uuid;
END;
$$ LANGUAGE PLPGSQL
VOLATILE;

-- from https://stackoverflow.com/questions/46433459/postgres-select-where-the-where-is-uuid-or-string/46433640#46433640
CREATE OR REPLACE FUNCTION dbfunc_schema.uuid_or_null(str text) RETURNS uuid
AS $$
BEGIN
  RETURN str::uuid;
EXCEPTION WHEN invalid_text_representation THEN
  RETURN NULL;
END;
$$ LANGUAGE PLPGSQL;

CREATE TABLE dbtable_schema.users (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  username VARCHAR (255) NOT NULL UNIQUE,
  sub uuid NOT NULL UNIQUE,
  image VARCHAR (250),
  first_name VARCHAR (255),
  last_name VARCHAR (255),
  email VARCHAR (255),
  ip_address VARCHAR (40),
  timezone VARCHAR (50),
  locked BOOLEAN NOT NULL DEFAULT false,
  active BOOLEAN NOT NULL DEFAULT false,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
CREATE INDEX user_sub_index ON dbtable_schema.users (sub);
ALTER TABLE dbtable_schema.users ENABLE ROW LEVEL SECURITY;
-- $IS_CREATOR checks allow a user to manage their own group user
-- The group user owns various records like schedules, forms, etc.
CREATE POLICY table_select ON dbtable_schema.users FOR SELECT TO $PG_WORKER USING ($IS_WORKER OR $IS_CREATOR OR $IS_USER);
CREATE POLICY table_insert ON dbtable_schema.users FOR INSERT TO $PG_WORKER WITH CHECK (true);
CREATE POLICY table_update ON dbtable_schema.users FOR UPDATE TO $PG_WORKER USING ($IS_CREATOR OR $IS_USER);
CREATE POLICY table_delete ON dbtable_schema.users FOR DELETE TO $PG_WORKER USING ($IS_CREATOR OR $IS_USER);

CREATE TABLE dbtable_schema.roles (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  name VARCHAR (50) NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.roles FOR SELECT TO $PG_WORKER USING (true);
CREATE POLICY table_insert ON dbtable_schema.roles FOR INSERT TO $PG_WORKER WITH CHECK (true);

DO $$
DECLARE admin_id uuid;
BEGIN
  admin_id := dbfunc_schema.uuid_generate_v7();
  INSERT INTO dbtable_schema.users (username, sub, created_sub) VALUES ('system_owner', admin_id, admin_id);
  INSERT INTO dbtable_schema.roles (name, created_sub) VALUES ('Admin', admin_id);
END $$;

CREATE TABLE dbtable_schema.file_types (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  name VARCHAR (50) NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.file_types ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.file_types FOR SELECT TO $PG_WORKER USING (true);

INSERT INTO
  dbtable_schema.file_types (name)
VALUES
  ('images'),
  ('documents');

CREATE TABLE dbtable_schema.files (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  uuid VARCHAR (50) NOT NULL,
  name VARCHAR (500) NOT NULL,
  mime_type TEXT,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.files ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.files FOR SELECT TO $PG_WORKER USING ($IS_CREATOR);
CREATE POLICY table_insert ON dbtable_schema.files FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_delete ON dbtable_schema.files FOR DELETE TO $PG_WORKER USING ($IS_CREATOR);

CREATE TABLE dbtable_schema.file_contents (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  uuid VARCHAR (50) NOT NULL,
  name VARCHAR (500) NOT NULL,
  content BYTEA NOT NULL,
  content_length INTEGER NOT NULL,
  upload_id VARCHAR (50) NOT NULL,
  expires_at TIMESTAMP NOT NULL DEFAULT NOW() + (60 * interval '1 day'),
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.file_contents ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.file_contents FOR SELECT TO $PG_WORKER USING ($IS_CREATOR);
CREATE POLICY table_insert ON dbtable_schema.file_contents FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_delete ON dbtable_schema.file_contents FOR DELETE TO $PG_WORKER USING ($IS_CREATOR);

CREATE TABLE dbtable_schema.groups (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  external_id uuid NOT NULL UNIQUE,
  admin_role_external_id TEXT NOT NULL UNIQUE,
  default_role_id uuid REFERENCES dbtable_schema.roles (id),
  display_name VARCHAR (100) NOT NULL UNIQUE,
  name VARCHAR (50) NOT NULL UNIQUE,
  purpose VARCHAR (200) NOT NULL,
  allowed_domains TEXT,
  code TEXT NOT NULL,
  ai BOOLEAN NOT NULL DEFAULT true,
  sub uuid NOT NULL REFERENCES dbtable_schema.users (sub) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.groups ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.groups FOR SELECT TO $PG_WORKER USING (
  $IS_WORKER OR $IS_CREATOR OR code IS NOT NULL
);
CREATE POLICY table_insert ON dbtable_schema.groups FOR INSERT TO $PG_WORKER WITH CHECK (
  NOT EXISTS(SELECT 1 FROM dbtable_schema.groups WHERE $IS_CREATOR)
);
CREATE POLICY table_update ON dbtable_schema.groups FOR UPDATE TO $PG_WORKER USING ($IS_CREATOR);
CREATE POLICY table_delete ON dbtable_schema.groups FOR DELETE TO $PG_WORKER USING ($IS_CREATOR);

CREATE TABLE dbtable_schema.group_roles (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  role_id uuid NOT NULL REFERENCES dbtable_schema.roles (id) ON DELETE CASCADE,
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  external_id TEXT NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (role_id, group_id)
);
ALTER TABLE dbtable_schema.group_roles ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_roles FOR SELECT TO $PG_WORKER USING (
  $IS_CREATOR OR $HAS_GROUP OR EXISTS (
    SELECT 1 FROM dbtable_schema.groups g 
    WHERE g.id = group_roles.group_id AND g.code IS NOT NULL
  )
);
CREATE POLICY table_insert ON dbtable_schema.group_roles FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR OR $HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_roles FOR UPDATE TO $PG_WORKER USING ($IS_CREATOR OR $HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_roles FOR DELETE TO $PG_WORKER USING ($IS_CREATOR OR $HAS_GROUP);

CREATE TABLE dbtable_schema.group_users (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  user_id uuid NOT NULL REFERENCES dbtable_schema.users (id) ON DELETE CASCADE,
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  external_id TEXT NOT NULL, -- this refers to the external subgroup id i.e. app db group role external id
  locked BOOLEAN NOT NULL DEFAULT false,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (user_id, group_id)
);
ALTER TABLE dbtable_schema.group_users ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_users FOR SELECT TO $PG_WORKER USING ($IS_CREATOR OR $HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_users FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR OR $HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_users FOR UPDATE TO $PG_WORKER USING ($IS_CREATOR OR $HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_users FOR DELETE TO $PG_WORKER USING ($IS_CREATOR OR $HAS_GROUP);

CREATE POLICY table_select_by_group_admin ON dbtable_schema.users FOR SELECT TO $PG_WORKER USING (
  EXISTS(
    SELECT 1 FROM dbtable_schema.group_users gu
    WHERE gu.user_id = dbtable_schema.users.id AND gu.$HAS_GROUP -- can select if the user record belongs to the session group id
  )
);

CREATE TABLE dbtable_schema.group_files (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  file_id uuid NOT NULL REFERENCES dbtable_schema.files (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (created_sub, file_id)
);
ALTER TABLE dbtable_schema.group_files ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_files FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_files FOR INSERT TO $PG_WORKER WITH CHECK ($HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_files FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_files FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.uuid_notes (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  parent_uuid VARCHAR (50) NOT NULL,
  note VARCHAR (500),
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (parent_uuid, note, created_sub)
);

CREATE TABLE dbtable_schema.request_log (
  id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
  sub VARCHAR (50) NOT NULL,
  path VARCHAR (500),
  direction VARCHAR (10),
  code VARCHAR (5),
  payload VARCHAR (5000),
  ip_address VARCHAR (50) NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

