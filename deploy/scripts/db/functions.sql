CREATE OR REPLACE FUNCTION dbfunc_schema.set_session_vars(
  p_user_sub VARCHAR,
  p_group_id VARCHAR,
  p_role_bits INTEGER DEFAULT 0,
  p_sock_topic VARCHAR DEFAULT ''
) RETURNS VOID AS $$
BEGIN
  PERFORM set_config('app_session.user_sub', p_user_sub, true);
  PERFORM set_config('app_session.group_id', p_group_id, true);
  PERFORM set_config('app_session.role_bits', p_role_bits::text, true);
  PERFORM set_config('app_session.sock_topic', p_sock_topic, true);
END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION dbfunc_schema.make_generic_code() RETURNS TRIGGER 
AS $$
DECLARE
  target_column text := TG_ARGV[0];
  code_length   int  := TG_ARGV[1];
  sql_query     text;
BEGIN
  sql_query := format(
    'UPDATE %I.%I SET %I = LOWER(SUBSTRING(MD5(''''||NOW()::TEXT||RANDOM()::TEXT) FOR %s)) WHERE "id" = $1.id',
    TG_TABLE_SCHEMA,
    TG_TABLE_NAME,
    target_column,
    code_length
  );

  LOOP
    BEGIN
      EXECUTE sql_query USING NEW;
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
  "partType" TEXT,
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
    'startTime', repslot.start_time
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

CREATE OR REPLACE FUNCTION dbfunc_schema.trg_handle_seat_payment()
RETURNS TRIGGER AS $$
BEGIN
  -- New Payment Created (Pending or Paid) -> Add Seats
  IF (TG_OP = 'INSERT') THEN
    -- Add the seats
    UPDATE dbtable_schema.group_seats
    SET balance = balance + NEW.seats, updated_sub = NEW.created_sub, updated_on = NOW()
    WHERE group_id = NEW.group_id;
  
  -- Payment Updated to VOID -> Remove Seats
  ELSIF (TG_OP = 'UPDATE') THEN
    -- If status of the seat_payments record changed to 'void' from something else, subtract the seats from group_seats
    IF (NEW.status = 'void' AND OLD.status != 'void') THEN
      UPDATE dbtable_schema.group_seats
      SET balance = balance - NEW.seats, updated_sub = NEW.created_sub, updated_on = NOW()
      WHERE group_id = NEW.group_id;
    END IF;
    
    -- un-void (manual correction), add them back
    IF (NEW.status != 'void' AND OLD.status = 'void') THEN
      UPDATE dbtable_schema.group_seats
      SET balance = balance + NEW.seats, updated_sub = NEW.created_sub, updated_on = NOW()
      WHERE group_id = NEW.group_id;
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION dbfunc_schema.register_monthly_seat_usage(
  p_group_id UUID, 
  p_user_sub UUID
)
RETURNS VOID AS $$
DECLARE
  v_month DATE := DATE_TRUNC('month', NOW());
BEGIN
  -- Try to insert the usage record
  INSERT INTO dbtable_schema.group_seat_usage (group_id, created_sub, month_date)
  VALUES (p_group_id, p_user_sub, v_month)
  ON CONFLICT (group_id, created_sub, month_date) DO NOTHING;

  -- If the insertion actually happened (row was created), decrement balance
  -- We check the system variable 'FOUND' to see if the INSERT processed a row
  IF FOUND THEN
    UPDATE dbtable_schema.group_seats
    SET balance = balance - 1, updated_sub = p_user_sub, updated_on = NOW()
    WHERE group_id = p_group_id;
  END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION dbfunc_schema.check_group_standing(p_group_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
  overdue_invoices INT;
  current_balance INT;
BEGIN
  -- Stop service if any unpaid invoice is older than 30 days
  SELECT COUNT(*) INTO overdue_invoices
  FROM dbtable_schema.seat_payments p
  WHERE p.group_id = p_group_id
  AND p.status = 'pending'
  AND p.created_on < NOW() - INTERVAL '30 days';

  IF overdue_invoices > 0 THEN
    RETURN FALSE;
  END IF;

  -- Check if the balance is negative
  SELECT balance INTO current_balance
  FROM dbtable_schema.group_seats
  WHERE group_id = p_group_id;

  -- If no row exists, balance is effectively 0 (or we treat as valid until used)
  -- If row exists and balance < 0, return false
  IF current_balance IS NOT NULL AND current_balance < 0 THEN
    RETURN FALSE;
  END IF;

  RETURN TRUE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
