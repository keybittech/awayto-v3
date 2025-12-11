CREATE TABLE dbtable_schema.budgets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (50) NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.budgets ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.budgets FOR SELECT TO $PG_WORKER USING (true);

INSERT INTO
  dbtable_schema.budgets (name)
VALUES
  ('$500 - $1,000'),
  ('$1,000 - $10,000'),
  ('$10,000 - $100,000');

CREATE TABLE dbtable_schema.timelines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (50) NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.timelines ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.timelines FOR SELECT TO $PG_WORKER USING (true);

INSERT INTO
  dbtable_schema.timelines (name)
VALUES
  ('1 month'),
  ('6 months'),
  ('1 year');

CREATE TABLE dbtable_schema.forms (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (500) NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.group_forms (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  form_id uuid NOT NULL REFERENCES dbtable_schema.forms (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub) ON DELETE CASCADE,
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (group_id, form_id)
);
ALTER TABLE dbtable_schema.group_forms ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_insert ON dbtable_schema.group_forms FOR INSERT TO $PG_WORKER WITH CHECK ($HAS_GROUP);
CREATE POLICY table_select ON dbtable_schema.group_forms FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_forms FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.form_versions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  form_id uuid NOT NULL REFERENCES dbtable_schema.forms (id) ON DELETE CASCADE,
  form JSONB NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.form_version_submissions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  form_version_id uuid NOT NULL REFERENCES dbtable_schema.form_versions (id) ON DELETE CASCADE,
  submission JSONB NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.services (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (50) NOT NULL,
  cost INTEGER,
  form_id uuid REFERENCES dbtable_schema.forms (id),
  survey_id uuid REFERENCES dbtable_schema.forms (id),
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (name, created_sub)
);

CREATE TABLE dbtable_schema.group_services (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES dbtable_schema.services (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (group_id, service_id)
);
ALTER TABLE dbtable_schema.group_services ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_services FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_services FOR INSERT TO $PG_WORKER WITH CHECK ($HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_services FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_services FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.service_addons (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (50) NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.group_service_addons (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  service_addon_id uuid NOT NULL REFERENCES dbtable_schema.service_addons (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (group_id, service_addon_id)
);

CREATE TABLE dbtable_schema.service_tiers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  service_id uuid NOT NULL REFERENCES dbtable_schema.services (id) ON DELETE CASCADE,
  form_id uuid REFERENCES dbtable_schema.forms (id),
  survey_id uuid REFERENCES dbtable_schema.forms (id),
  name VARCHAR (500) NOT NULL,
  multiplier INTEGER NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (name, service_id)
);

CREATE TABLE dbtable_schema.service_tier_addons (
  service_tier_id uuid NOT NULL REFERENCES dbtable_schema.service_tiers (id) ON DELETE CASCADE,
  service_addon_id uuid NOT NULL REFERENCES dbtable_schema.service_addons (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (service_tier_id, service_addon_id)
);

CREATE TABLE dbtable_schema.time_units (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (50) NOT NULL UNIQUE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO
  dbtable_schema.time_units (name)
VALUES
  ('minute'),
  ('hour'),
  ('day'),
  ('week'),
  ('month'),
  ('year');

CREATE TABLE dbtable_schema.schedules (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR (50),
  start_date TIMESTAMPTZ,
  end_date TIMESTAMPTZ,
  timezone VARCHAR(128) NOT NULL,
  schedule_time_unit_id uuid NOT NULL REFERENCES dbtable_schema.time_units (id),
  bracket_time_unit_id uuid NOT NULL REFERENCES dbtable_schema.time_units (id),
  slot_time_unit_id uuid NOT NULL REFERENCES dbtable_schema.time_units (id),
  slot_duration INTEGER NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub) ON DELETE CASCADE,
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.group_schedules ( -- master schedules tied to the group, created sub is group db user
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  schedule_id uuid NOT NULL REFERENCES dbtable_schema.schedules (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub) ON DELETE CASCADE,
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (group_id, schedule_id)
);
ALTER TABLE dbtable_schema.group_schedules ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_schedules FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_schedules FOR INSERT TO $PG_WORKER WITH CHECK ($HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_schedules FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_schedules FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.group_user_schedules ( -- user schedules based off the masters
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  group_schedule_id uuid NOT NULL REFERENCES dbtable_schema.schedules (id) ON DELETE CASCADE,
  user_schedule_id uuid NOT NULL REFERENCES dbtable_schema.schedules (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (group_schedule_id, user_schedule_id)
);
ALTER TABLE dbtable_schema.group_user_schedules ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_user_schedules FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_user_schedules FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR AND $HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_user_schedules FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_user_schedules FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

-- putting the schedule rules here, after group_schedules and group_user_schedules
-- allows for is_creator (regular users and group db users) to edit their own schedules
-- and then group_schedules/group_user_schedules relationships allow group member access
ALTER TABLE dbtable_schema.schedules ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.schedules FOR SELECT TO $PG_WORKER USING (
  $IS_CREATOR
  -- group members can act on master schedules
  OR EXISTS(SELECT 1 FROM dbtable_schema.group_schedules gs WHERE gs.schedule_id = dbtable_schema.schedules.id AND gs.$HAS_GROUP)
  -- group members can act on other user schedules
  OR EXISTS(SELECT 1 FROM dbtable_schema.group_user_schedules gus WHERE gus.user_schedule_id = dbtable_schema.schedules.id AND gus.$HAS_GROUP)
);
CREATE POLICY table_insert ON dbtable_schema.schedules FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.schedules FOR UPDATE TO $PG_WORKER USING (
  $IS_CREATOR
  OR EXISTS(SELECT 1 FROM dbtable_schema.group_schedules gs WHERE gs.schedule_id = dbtable_schema.schedules.id AND gs.$HAS_GROUP)
  OR EXISTS(SELECT 1 FROM dbtable_schema.group_user_schedules gus WHERE gus.user_schedule_id = dbtable_schema.schedules.id AND gus.$HAS_GROUP)
);
CREATE POLICY table_delete ON dbtable_schema.schedules FOR DELETE TO $PG_WORKER USING (
  $IS_CREATOR
  OR EXISTS(SELECT 1 FROM dbtable_schema.group_schedules gs WHERE gs.schedule_id = dbtable_schema.schedules.id AND gs.$HAS_GROUP)
  OR EXISTS(SELECT 1 FROM dbtable_schema.group_user_schedules gus WHERE gus.user_schedule_id = dbtable_schema.schedules.id AND gus.$HAS_GROUP)
);

CREATE TABLE dbtable_schema.schedule_brackets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  schedule_id uuid NOT NULL REFERENCES dbtable_schema.schedules (id) ON DELETE CASCADE,
  duration INTEGER NOT NULL,
  multiplier INTEGER NOT NULL,
  automatic BOOLEAN NOT NULL DEFAULT false,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.schedule_brackets ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.schedule_brackets FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.schedule_brackets FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.schedule_brackets FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.schedule_brackets FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.schedule_bracket_slots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  schedule_bracket_id uuid NOT NULL REFERENCES dbtable_schema.schedule_brackets (id) ON DELETE CASCADE,
  start_time INTERVAL NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.schedule_bracket_slots ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.schedule_bracket_slots FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.schedule_bracket_slots FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.schedule_bracket_slots FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.schedule_bracket_slots FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE UNIQUE INDEX idx_unique_enabled_slots 
ON dbtable_schema.schedule_bracket_slots(schedule_bracket_id, start_time) 
WHERE enabled = true;

CREATE TABLE dbtable_schema.schedule_bracket_slot_exclusions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  exclusion_date DATE NOT NULL,
  schedule_bracket_slot_id uuid REFERENCES dbtable_schema.schedule_bracket_slots (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.schedule_bracket_slot_exclusions ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.schedule_bracket_slot_exclusions FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.schedule_bracket_slot_exclusions FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.schedule_bracket_slot_exclusions FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.schedule_bracket_slot_exclusions FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.schedule_bracket_services (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  schedule_bracket_id uuid NOT NULL REFERENCES dbtable_schema.schedule_brackets (id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES dbtable_schema.services (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE (schedule_bracket_id, service_id)
);
ALTER TABLE dbtable_schema.schedule_bracket_services ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.schedule_bracket_services FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.schedule_bracket_services FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.schedule_bracket_services FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.schedule_bracket_services FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TABLE dbtable_schema.quotes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  slot_date DATE NOT NULL,
  schedule_bracket_slot_id uuid NOT NULL REFERENCES dbtable_schema.schedule_bracket_slots (id),
  service_tier_id uuid NOT NULL REFERENCES dbtable_schema.service_tiers (id),
  service_form_version_submission_id uuid REFERENCES dbtable_schema.form_version_submissions (id),
  tier_form_version_submission_id uuid REFERENCES dbtable_schema.form_version_submissions (id),
  slot_created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.quotes ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.quotes FOR SELECT TO $PG_WORKER USING (
  $IS_CREATOR OR EXISTS(
    SELECT 1 FROM dbtable_schema.schedule_bracket_slots sbs WHERE sbs.id = dbtable_schema.quotes.schedule_bracket_slot_id AND sbs.$IS_CREATOR
  )
);
CREATE POLICY table_insert ON dbtable_schema.quotes FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.quotes FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);

CREATE POLICY table_select_2 ON dbtable_schema.files FOR SELECT TO $PG_WORKER USING (
  EXISTS( -- this allows staff members to see information about the users they have appointments with
    SELECT 1 FROM dbtable_schema.quotes q
    JOIN dbtable_schema.schedule_bracket_slots sbs ON q.schedule_bracket_slot_id = sbs.id
    WHERE q.created_sub = dbtable_schema.files.created_sub -- selecting record in question belongs to user who made the quote
    AND sbs.$IS_CREATOR -- selecting user made the schedule bracket for the quote
  )
);

CREATE POLICY table_select_2 ON dbtable_schema.file_contents FOR SELECT TO $PG_WORKER USING (
  EXISTS( -- this allows staff members to see information about the users they have appointments with
    SELECT 1 FROM dbtable_schema.quotes q
    JOIN dbtable_schema.schedule_bracket_slots sbs ON q.schedule_bracket_slot_id = sbs.id
    WHERE q.created_sub = dbtable_schema.file_contents.created_sub -- selecting record in question belongs to user who made the quote
    AND sbs.$IS_CREATOR -- selecting user made the schedule bracket for the quote
  )
);

CREATE TABLE dbtable_schema.quote_files (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_id uuid NOT NULL REFERENCES dbtable_schema.quotes (id) ON DELETE CASCADE,
  file_id uuid NOT NULL REFERENCES dbtable_schema.files (id) ON DELETE CASCADE,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.quote_files ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.quote_files FOR SELECT TO $PG_WORKER USING (
  $IS_CREATOR OR EXISTS(
    SELECT 1 FROM dbtable_schema.quotes q WHERE q.id = dbtable_schema.quote_files.quote_id AND q.$IS_CREATOR
  )
);
CREATE POLICY table_insert ON dbtable_schema.quote_files FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);

CREATE TABLE dbtable_schema.bookings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_id uuid NOT NULL REFERENCES dbtable_schema.quotes (id),
  slot_date DATE NOT NULL,
  schedule_bracket_slot_id uuid NOT NULL REFERENCES dbtable_schema.schedule_bracket_slots (id),
  service_survey_version_submission_id uuid REFERENCES dbtable_schema.form_version_submissions (id),
  tier_survey_version_submission_id uuid REFERENCES dbtable_schema.form_version_submissions (id),
  rating SMALLINT,
  quote_created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.bookings ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.bookings FOR SELECT TO $PG_WORKER USING (
  $IS_CREATOR OR EXISTS(
    SELECT 1 FROM dbtable_schema.quotes q WHERE q.id = dbtable_schema.bookings.quote_id AND q.$IS_CREATOR
  )
);
CREATE POLICY table_insert ON dbtable_schema.bookings FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.bookings FOR UPDATE TO $PG_WORKER USING ($IS_CREATOR);

CREATE TABLE dbtable_schema.sock_connections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id TEXT NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub) ON DELETE CASCADE,
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.sock_connections ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.sock_connections FOR SELECT TO $PG_WORKER USING ($IS_WORKER OR $IS_CREATOR OR
  EXISTS( -- this allows staff members to see information about the users they have appointments with
    SELECT 1 FROM dbtable_schema.quotes q
    JOIN dbtable_schema.schedule_bracket_slots sbs ON q.schedule_bracket_slot_id = sbs.id
    WHERE q.created_sub = dbtable_schema.sock_connections.created_sub -- selecting record in question belongs to user who made the quote
    AND sbs.$IS_CREATOR -- selecting user made the schedule bracket for the quote
  )
);
CREATE POLICY table_insert ON dbtable_schema.sock_connections FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_delete ON dbtable_schema.sock_connections FOR DELETE TO $PG_WORKER USING ($IS_WORKER);

CREATE TABLE dbtable_schema.topic_messages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id TEXT NOT NULL,
  topic TEXT NOT NULL,
  message TEXT NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.topic_messages ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.topic_messages FOR SELECT TO $PG_WORKER USING (
  $IS_WORKER OR dbtable_schema.topic_messages.topic = current_setting('app_session.sock_topic')
);
CREATE POLICY table_insert ON dbtable_schema.topic_messages FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE INDEX topic_index ON dbtable_schema.topic_messages (topic);

CREATE TABLE dbtable_schema.topic_canvas_elements (
  connection_id TEXT NOT NULL,
  topic TEXT NOT NULL,
  element_id TEXT NOT NULL,
  element_type TEXT NOT NULL DEFAULT 'box',
  properties JSONB NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  PRIMARY KEY (topic, element_id)
);
ALTER TABLE dbtable_schema.topic_canvas_elements ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select_state ON dbtable_schema.topic_canvas_elements FOR SELECT TO $PG_WORKER USING (
  $IS_WORKER OR dbtable_schema.topic_canvas_elements.topic = current_setting('app_session.sock_topic')
);
CREATE POLICY table_modify_state ON dbtable_schema.topic_canvas_elements FOR ALL TO $PG_WORKER USING ($IS_CREATOR);
CREATE INDEX topic_element_index ON dbtable_schema.topic_canvas_elements (topic, element_id);

CREATE TABLE dbtable_schema.exchange_call_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id uuid NOT NULL REFERENCES dbtable_schema.bookings (id),
  style TEXT NOT NULL,
  connected TIMESTAMP NOT NULL,
  disconnected TIMESTAMP,
  transcript JSONB, -- this denotes audio based chat logs
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.feedback (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  message TEXT,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE dbtable_schema.group_feedback (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  message TEXT,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.group_feedback ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_feedback FOR SELECT TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_feedback FOR INSERT TO $PG_WORKER WITH CHECK ($HAS_GROUP);
CREATE POLICY table_update ON dbtable_schema.group_feedback FOR UPDATE TO $PG_WORKER USING ($HAS_GROUP);
CREATE POLICY table_delete ON dbtable_schema.group_feedback FOR DELETE TO $PG_WORKER USING ($HAS_GROUP);

CREATE TYPE dbtable_schema.payment_status AS ENUM ('pending', 'paid', 'void');

CREATE TABLE dbtable_schema.seat_payments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id),
  status dbtable_schema.payment_status DEFAULT 'pending',
  code TEXT NOT NULL,
  seats INTEGER NOT NULL,
  amount INTEGER NOT NULL,
  paid_on TIMESTAMP,
  check_no TEXT,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.seat_payments ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.seat_payments FOR SELECT TO $PG_WORKER USING ($IS_WORKER OR $HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.seat_payments FOR INSERT TO $PG_WORKER WITH CHECK ($IS_WORKER);
CREATE POLICY table_update ON dbtable_schema.seat_payments FOR UPDATE TO $PG_WORKER USING ($IS_WORKER);

CREATE TABLE dbtable_schema.group_seats (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE UNIQUE,
  balance SMALLINT NOT NULL DEFAULT 5,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub),
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true
);
ALTER TABLE dbtable_schema.group_seats ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_seats FOR SELECT TO $PG_WORKER USING ($IS_WORKER OR $HAS_GROUP);
CREATE POLICY table_insert ON dbtable_schema.group_seats FOR INSERT TO $PG_WORKER WITH CHECK ($IS_CREATOR);
CREATE POLICY table_update ON dbtable_schema.group_seats FOR UPDATE TO $PG_WORKER USING ($IS_WORKER);
CREATE POLICY table_delete ON dbtable_schema.group_seats FOR DELETE TO $PG_WORKER USING ($IS_WORKER);

CREATE TABLE dbtable_schema.group_seat_usage (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id uuid NOT NULL REFERENCES dbtable_schema.groups (id) ON DELETE CASCADE,
  month_date DATE NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT TIMEZONE('utc', NOW()),
  created_sub uuid NOT NULL REFERENCES dbtable_schema.users (sub) ON DELETE CASCADE,
  updated_on TIMESTAMP,
  updated_sub uuid REFERENCES dbtable_schema.users (sub),
  enabled BOOLEAN NOT NULL DEFAULT true,
  UNIQUE(group_id, created_sub, month_date)
);
ALTER TABLE dbtable_schema.group_seat_usage ENABLE ROW LEVEL SECURITY;
CREATE POLICY table_select ON dbtable_schema.group_seat_usage FOR SELECT TO $PG_WORKER USING ($IS_WORKER);
CREATE POLICY table_insert ON dbtable_schema.group_seat_usage FOR INSERT TO $PG_WORKER WITH CHECK ($IS_WORKER);
