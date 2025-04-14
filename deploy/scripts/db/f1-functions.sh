#!/bin/bash

psql -v ON_ERROR_STOP=1 --dbname $PG_DB <<-EOSQL

  CREATE OR REPLACE FUNCTION dbfunc_schema.set_session_vars(
    p_user_sub VARCHAR,
    p_group_id VARCHAR,
    p_roles VARCHAR
  ) RETURNS VOID AS \$\$
  BEGIN

    EXECUTE format('SET SESSION app_session.user_sub = %L', p_user_sub);
    EXECUTE format('SET SESSION app_session.group_id = %L', p_group_id);
    EXECUTE format('SET SESSION app_session.roles = %L', p_roles);
  
  END;
  \$\$ LANGUAGE PLPGSQL;

  CREATE OR REPLACE FUNCTION dbfunc_schema.delete_group(
    sub UUID  
  ) RETURNS TABLE (
    id UUID
  ) AS \$\$
  BEGIN

    RETURN QUERY
    DELETE FROM dbtable_schema.groups WHERE created_sub = sub
    RETURNING dbtable_schema.groups.id;

  END;
  \$\$ LANGUAGE PLPGSQL;

  CREATE OR REPLACE FUNCTION dbfunc_schema.get_group_schedules(
    p_month_start_date DATE, 
    p_schedule_id UUID, 
    p_client_timezone TEXT
  ) RETURNS TABLE (
    "weekStart" TEXT,
    "startDate" TEXT,
    "startTime" TEXT,
    "scheduleBracketSlotId" UUID
  )  AS \$\$
  BEGIN
    RETURN QUERY
    WITH series AS (
      SELECT
        week_start::DATE,
        slot."startTime"::INTERVAL as start_time,
        slot.id as schedule_bracket_slot_id,
        schedule.timezone as schedule_timezone
      FROM generate_series(
        DATE_TRUNC('week', p_month_start_date::DATE + INTERVAL '1 day'),
        DATE_TRUNC('week', p_month_start_date::DATE + INTERVAL '1 day') + INTERVAL '1 month',
        interval '1 week'
      ) AS week_start
      CROSS JOIN dbview_schema.enabled_schedule_bracket_slots slot
      LEFT JOIN dbtable_schema.bookings booking ON booking.schedule_bracket_slot_id = slot.id AND DATE_TRUNC('week', booking.slot_date + INTERVAL '1 day') - INTERVAL '1 DAY' = week_start
      LEFT JOIN dbtable_schema.schedule_bracket_slot_exclusions exclusion ON exclusion.schedule_bracket_slot_id = slot.id AND DATE_TRUNC('week', exclusion.exclusion_date + INTERVAL '1 day') - INTERVAL '1 DAY' = week_start
      JOIN dbtable_schema.schedule_brackets bracket ON bracket.id = slot."scheduleBracketId"
      JOIN dbtable_schema.group_user_schedules gus ON gus.user_schedule_id = bracket.schedule_id
      JOIN dbtable_schema.schedules schedule ON schedule.id = gus.group_schedule_id
      WHERE
        booking.id IS NULL AND
        exclusion.id IS NULL AND
        schedule.id = p_schedule_id
    ), timers AS (
      SELECT
        series.*,
        (week_start + start_time) AT TIME ZONE schedule_timezone AS schedule_time,
        (week_start + start_time) AT TIME ZONE schedule_timezone AT TIME ZONE p_client_timezone AS client_time,
        EXTRACT(EPOCH FROM (week_start + start_time) AT TIME ZONE schedule_timezone AT TIME ZONE p_client_timezone - DATE_TRUNC('week', (week_start + start_time) AT TIME ZONE schedule_timezone AT TIME ZONE p_client_timezone)) AS seconds_from_week_start
      FROM
        series
    )
    SELECT 
      TO_CHAR(DATE_TRUNC('week', client_time)::DATE, 'YYYY-MM-DD')::TEXT as "weekStart",
      TO_CHAR(client_time::DATE, 'YYYY-MM-DD')::TEXT as "startDate",
      CONCAT(
        'P',
        FLOOR(seconds_from_week_start / 86400)::TEXT, 'DT',
        FLOOR((seconds_from_week_start % 86400) / 3600)::TEXT, 'H',
        FLOOR((seconds_from_week_start % 3600) / 60)::TEXT, 'M'
      ) as "startTime",
      schedule_bracket_slot_id as "scheduleBracketSlotId"
    FROM timers
    WHERE 
      schedule_time > (NOW() AT TIME ZONE schedule_timezone) AND
      schedule_time <> schedule_time + INTERVAL '1 hour'
    ORDER BY client_time;
  END;
  \$\$ LANGUAGE PLPGSQL;

  CREATE OR REPLACE FUNCTION dbfunc_schema.make_group_code() RETURNS TRIGGER 
  AS \$\$
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
  \$\$ LANGUAGE PLPGSQL VOLATILE;

  CREATE OR REPLACE FUNCTION dbfunc_schema.get_scheduled_parts (
    p_schedule_id UUID
  )
  RETURNS TABLE (
    partType TEXT,
    ids JSONB
  )  AS \$\$
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
  \$\$ LANGUAGE PLPGSQL;

  CREATE OR REPLACE FUNCTION dbfunc_schema.get_peer_schedule_replacement (
    p_user_schedule_ids UUID[],
    p_slot_date DATE,
    p_start_time INTERVAL,
    p_tier_name TEXT
  )
  RETURNS TABLE (
    replacement JSON
  )  AS \$\$
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
  \$\$ LANGUAGE PLPGSQL;

  CREATE FUNCTION dbfunc_schema.is_slot_taken(p_slot_id uuid, p_date date) 
  RETURNS boolean AS \$\$
  BEGIN
    RETURN EXISTS (
      SELECT 1 FROM dbtable_schema.bookings
      WHERE schedule_bracket_slot_id = p_slot_id 
      AND slot_date = p_date
    );
  END;
  \$\$ LANGUAGE plpgsql SECURITY DEFINER;
EOSQL
