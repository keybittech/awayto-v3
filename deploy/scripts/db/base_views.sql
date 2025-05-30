CREATE
OR REPLACE VIEW dbview_schema.enabled_users
AS
SELECT
  u.id,
  u.first_name as "firstName",
  u.last_name as "lastName",
  u.sub,
  u.image,
  u.email,
  u.locked,
  u.active,
  u.created_on as "createdOn",
  u.updated_on as "updatedOn",
  u.enabled
FROM
  dbtable_schema.users u
WHERE
  u.enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_roles AS
SELECT
  id,
  name,
  created_on as "createdOn"
FROM
  dbtable_schema.roles
WHERE
  enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_groups AS
SELECT
  id,
  name,
  sub,
  code,
  COALESCE(default_role_id::TEXT, '') as "defaultRoleId",
  allowed_domains as "allowedDomains",
  external_id as "externalId",
  display_name as "displayName",
  created_sub as "createdSub",
  created_on as "createdOn",
  purpose,
  ai
FROM
  dbtable_schema.groups
WHERE
  enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_group_users AS
SELECT
  id,
  user_id as "userId",
  group_id as "groupId",
  created_on as "createdOn"
FROM
  dbtable_schema.group_users
WHERE
  enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_group_roles AS
SELECT
  id,
  role_id as "roleId",
  group_id as "groupId",
  external_id as "externalId",
  created_on as "createdOn"
FROM
  dbtable_schema.group_roles
WHERE
  enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_file_types AS
SELECT
  id,
  name,
  created_on as "createdOn"
FROM
  dbtable_schema.file_types
WHERE
  enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_files AS
SELECT
  f.id,
  f.uuid,
  f.name,
  f.mime_type as "mimeType",
  f.created_sub as "createdSub",
  f.created_on as "createdOn"
FROM
  dbtable_schema.files f
WHERE
  f.enabled = true;

CREATE
OR REPLACE VIEW dbview_schema.enabled_group_files AS
SELECT
  gf.id,
  gf.file_id as "fileId",
  f.name,
  gf.group_id as "groupId",
  gf.created_on as "createdOn"
FROM
  dbtable_schema.group_files gf
  JOIN dbview_schema.enabled_files f ON f.id = gf.file_id
WHERE
  gf.enabled = true;

