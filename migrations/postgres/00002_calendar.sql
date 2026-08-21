-- +goose Up
CREATE FUNCTION public.calendar_event_uid(
    group_number text,
    lesson_date date,
    lesson_start timestamp without time zone,
    lesson_title text,
    lesson_type text
) RETURNS text
LANGUAGE sql
STABLE
RETURN 'schedule-rsreu-' || md5(concat_ws('|',
    upper(group_number),
    to_char(lesson_date, 'YYYY-MM-DD'),
    to_char(lesson_start, 'HH24:MI:SS'),
    lesson_title,
    coalesce(lesson_type, '')
)) || '@rsreu-schedule.ru';

CREATE TABLE public.calendar_revision (
    id smallint PRIMARY KEY CHECK (id = 1),
    revision bigint NOT NULL DEFAULT 0,
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

INSERT INTO public.calendar_revision (id) VALUES (1);

CREATE TABLE public.calendar_deleted_event (
    uid text PRIMARY KEY,
    group_number text NOT NULL,
    start_time timestamp without time zone NOT NULL,
    end_time timestamp without time zone NOT NULL,
    title text NOT NULL,
    lesson_type text,
    teacher_auditoriums jsonb NOT NULL DEFAULT '[]'::jsonb,
    sequence bigint NOT NULL,
    cancelled_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX calendar_deleted_event_group_start_idx
    ON public.calendar_deleted_event (group_number, start_time);

-- +goose Down
DROP TABLE public.calendar_deleted_event;
DROP TABLE public.calendar_revision;
DROP FUNCTION public.calendar_event_uid(text, date, timestamp without time zone, text, text);
