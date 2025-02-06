#!/bin/bash

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-'EOSQL'


  DROP SCHEMA IF EXISTS dbfunc_schema CASCADE;
  CREATE SCHEMA dbfunc_schema;

  -- from https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
  create or replace function dbfunc_schema.uuid_generate_v7()
  returns uuid
  as $$
  begin
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
  end
  $$
  language plpgsql
  volatile;


  DROP SCHEMA IF EXISTS dbtable_schema CASCADE;
  CREATE SCHEMA dbtable_schema;

  CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

  CREATE TABLE dbtable_schema.users (
    id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
    username VARCHAR (255) NOT NULL UNIQUE,
    sub uuid NOT NULL UNIQUE,
    image VARCHAR (250),
    first_name VARCHAR (255),
    last_name VARCHAR (255),
    email VARCHAR (255),
    ip_address VARCHAR (40),
    locked BOOLEAN NOT NULL DEFAULT false,
    active BOOLEAN NOT NULL DEFAULT false,
    created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
    created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
    updated_on TIMESTAMP,
    updated_sub uuid REFERENCES dbtable_schema.users (sub),
    enabled BOOLEAN NOT NULL DEFAULT true
  );

  CREATE INDEX user_sub_index ON dbtable_schema.users (sub);

  CREATE TABLE dbtable_schema.roles (
    id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
    name VARCHAR (50) NOT NULL UNIQUE,
    created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
    created_sub uuid REFERENCES dbtable_schema.users (sub),
    updated_on TIMESTAMP,
    updated_sub uuid REFERENCES dbtable_schema.users (sub),
    enabled BOOLEAN NOT NULL DEFAULT true
  );

  CREATE TABLE dbtable_schema.user_roles (
    id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
    role_id uuid NOT NULL REFERENCES dbtable_schema.roles (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES dbtable_schema.users (id) ON DELETE CASCADE,
    created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
    created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
    updated_on TIMESTAMP,
    updated_sub uuid REFERENCES dbtable_schema.users (sub),
    enabled BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (role_id, user_id)
  );

  CREATE TABLE dbtable_schema.file_types (
    id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
    name VARCHAR (50) NOT NULL UNIQUE,
    created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
    created_sub uuid REFERENCES dbtable_schema.users (sub),
    updated_on TIMESTAMP,
    updated_sub uuid REFERENCES dbtable_schema.users (sub),
    enabled BOOLEAN NOT NULL DEFAULT true
  );

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

  CREATE TABLE dbtable_schema.file_contents (
    id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
    uuid VARCHAR (50) NOT NULL,
    name VARCHAR (500) NOT NULL,
    content BYTEA NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()) + (60 * interval '1 day'),
    created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
    created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
    updated_on TIMESTAMP,
    updated_sub uuid REFERENCES dbtable_schema.users (sub),
    enabled BOOLEAN NOT NULL DEFAULT true
  );

  CREATE TABLE dbtable_schema.groups (
    id uuid PRIMARY KEY DEFAULT dbfunc_schema.uuid_generate_v7(),
    external_id TEXT NOT NULL UNIQUE,
    admin_external_id TEXT NOT NULL UNIQUE,
    default_role_id uuid REFERENCES dbtable_schema.roles (id) ON DELETE CASCADE,
    display_name VARCHAR (100) NOT NULL UNIQUE,
    name VARCHAR (50) NOT NULL UNIQUE,
    purpose VARCHAR (200) NOT NULL,
    allowed_domains TEXT,
    code TEXT NOT NULL,
    ai BOOLEAN NOT NULL DEFAULT true,
    created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
    created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
    updated_on TIMESTAMP,
    updated_sub uuid REFERENCES dbtable_schema.users (sub),
    enabled BOOLEAN NOT NULL DEFAULT true
  );

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

EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  GRANT ALL ON SCHEMA dbtable_schema TO $POSTGRES_USER;
EOSQL
