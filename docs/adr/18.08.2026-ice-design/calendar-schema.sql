CREATE FUNCTION calendar_event_uid(
    group_number text,
    lesson_date date,
    lesson_start timestamp without time zone,
    lesson_title text,
    lesson_type text
) RETURNS text
LANGUAGE sql
IMMUTABLE
RETURN 'schedule-rsreu-' || md5(concat_ws('|',
    upper(group_number),
    lesson_date::text,
    lesson_start::time::text,
    lesson_title,
    coalesce(lesson_type, '')
)) || '@rsreu-schedule.ru';

CREATE TABLE calendar_revision (
    id smallint PRIMARY KEY CHECK (id = 1),
    revision bigint NOT NULL DEFAULT 0,
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

INSERT INTO calendar_revision (id) VALUES (1);

CREATE TABLE calendar_deleted_event (
    uid text PRIMARY KEY,
    group_number text NOT NULL,
    start_time timestamp without time zone NOT NULL,
    end_time timestamp without time zone NOT NULL,
    title text NOT NULL,
    lesson_type text,
    teachers text[] NOT NULL DEFAULT '{}',
    auditoriums text[] NOT NULL DEFAULT '{}',
    sequence bigint NOT NULL,
    cancelled_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX calendar_deleted_event_group_start_idx
    ON calendar_deleted_event (group_number, start_time);
