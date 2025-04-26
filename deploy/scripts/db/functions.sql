CREATE OR REPLACE FUNCTION dbfunc_schema.set_session_vars(
  p_user_sub VARCHAR,
  p_group_id VARCHAR,
  p_role_bits INTEGER DEFAULT 0,
  p_sock_topic VARCHAR DEFAULT ''
) RETURNS VOID AS $$
BEGIN
  EXECUTE format('SET SESSION app_session.user_sub = %L', p_user_sub);
  EXECUTE format('SET SESSION app_session.group_id = %L', p_group_id);
  EXECUTE format('SET SESSION app_session.role_bits = %L', p_role_bits);
  EXECUTE format('SET SESSION app_session.sock_topic = %L', p_sock_topic);
END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION dbfunc_schema.delete_group(
  sub UUID  
) RETURNS TABLE (
  id UUID
) AS $$
BEGIN
  RETURN QUERY
  DELETE FROM dbtable_schema.groups WHERE created_sub = sub
  RETURNING dbtable_schema.groups.id;
END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION dbfunc_schema.make_group_code() RETURNS TRIGGER 
AS $$
  BEGIN
    LOOP
      BEGIN
        UPDATE dbtable_schema.groups SET "code" = LOWER(SUBSTRING(MD5(''||NOW()::TEXT||RANDOM()::TEXT) FOR 8))
        WHERE "id" = NEW.id;
        EXIT;
      EXCEPTION WHEN unique_violation THEN
      END;
    END LOOP;
    RETURN NEW;
  END;
$$ LANGUAGE PLPGSQL VOLATILE;

CREATE OR REPLACE FUNCTION dbfunc_schema.get_scheduled_parts (
  p_schedule_id UUID
)
RETURNS TABLE (
  partType TEXT,
  ids JSONB
)  AS $$
BEGIN
RETURN QUERY
  SELECT DISTINCT 'slot', COALESCE(NULLIF(JSONB_AGG(sl.id), '[]'), '[]')
  FROM dbtable_schema.schedule_bracket_slots sl
  JOIN dbtable_schema.schedule_brackets bracket ON bracket.id = sl.schedule_bracket_id
  JOIN dbtable_schema.schedules schedule ON schedule.id = bracket.schedule_id
  LEFT JOIN dbtable_schema.quotes quote ON quote.schedule_bracket_slot_id = sl.id
  WHERE schedule.id = p_schedule_id AND quote.id IS NOT NULL
  UNION
  SELECT DISTINCT 'service', COALESCE(NULLIF(JSONB_AGG(se.id), '[]'), '[]')
  FROM dbtable_schema.schedule_bracket_services se
  JOIN dbtable_schema.schedule_brackets bracket ON bracket.id = se.schedule_bracket_id
  JOIN dbtable_schema.schedules schedule ON schedule.id = bracket.schedule_id
  JOIN dbtable_schema.services service ON service.id = se.service_id
  JOIN dbtable_schema.service_tiers tier ON tier.service_id = service.id
  LEFT JOIN dbtable_schema.quotes quote ON quote.service_tier_id = tier.id
  JOIN dbtable_schema.schedule_bracket_slots slot ON slot.id = quote.schedule_bracket_slot_id 
  WHERE schedule.id = p_schedule_id
  AND slot.schedule_bracket_id = bracket.id
  AND quote.id IS NOT NULL;
END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION dbfunc_schema.get_peer_schedule_replacement (
  p_user_schedule_ids UUID[],
  p_slot_date DATE,
  p_start_time INTERVAL,
  p_tier_name TEXT
)
RETURNS TABLE (
  replacement JSON
)  AS $$
BEGIN
RETURN QUERY
  SELECT JSON_BUILD_OBJECT(
    'username', usr.username,
    'scheduleBracketSlotId', repslot.id,
    'startTime', repslot.start_time::TEXT,
    'serviceTierId', reptier.id
  ) as replacement
  FROM
    dbtable_schema.group_user_schedules user_sched
  JOIN dbtable_schema.group_user_schedules peer_sched ON peer_sched.group_schedule_id = user_sched.group_schedule_id AND peer_sched.user_schedule_id <> ALL(p_user_schedule_ids::UUID[])
  JOIN dbtable_schema.schedule_brackets repbrac ON repbrac.schedule_id = peer_sched.user_schedule_id
  JOIN dbtable_schema.schedule_bracket_slots repslot ON repslot.schedule_bracket_id = repbrac.id
  JOIN dbtable_schema.schedule_bracket_services repserv ON repserv.schedule_bracket_id = repbrac.id
  JOIN dbtable_schema.service_tiers reptier ON reptier.service_id = repserv.service_id
  JOIN dbtable_schema.users usr ON usr.sub = repslot.created_sub
  LEFT JOIN dbtable_schema.quotes repq ON repq.schedule_bracket_slot_id = repslot.id AND repq.slot_date = p_slot_date AND repq.enabled = true
  WHERE user_sched.user_schedule_id = ANY(p_user_schedule_ids::UUID[])
    AND repslot.start_time = p_start_time
    AND reptier.name = p_tier_name
    AND repq.id IS NULL
    AND repslot.enabled = true
  LIMIT 1;
END;
$$ LANGUAGE PLPGSQL;

CREATE FUNCTION dbfunc_schema.is_slot_taken(p_slot_id uuid, p_date date) 
RETURNS boolean AS $$
BEGIN
  RETURN EXISTS (
    SELECT 1 FROM dbtable_schema.bookings
    WHERE schedule_bracket_slot_id = p_slot_id 
    AND slot_date = p_date
  );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
