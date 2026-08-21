-- +goose Up
ALTER TABLE public.calendar_deleted_event
ADD COLUMN teacher_auditoriums jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE public.calendar_deleted_event
SET teacher_auditoriums = (
    SELECT coalesce(jsonb_agg(item), '[]'::jsonb)
    FROM (
        SELECT jsonb_build_object(
            'teacher', teacher_name,
            'auditorium', ''
        ) AS item
        FROM unnest(teachers) AS teacher_name
        WHERE btrim(teacher_name) <> ''

        UNION

        SELECT jsonb_build_object(
            'teacher', '',
            'auditorium', auditorium_name
        ) AS item
        FROM unnest(auditoriums) AS auditorium_name
        WHERE btrim(auditorium_name) <> ''
    ) snapshot_items
);

ALTER TABLE public.calendar_deleted_event
DROP COLUMN teachers,
DROP COLUMN auditoriums;

-- +goose Down
ALTER TABLE public.calendar_deleted_event
ADD COLUMN teachers text[] NOT NULL DEFAULT '{}',
ADD COLUMN auditoriums text[] NOT NULL DEFAULT '{}';

UPDATE public.calendar_deleted_event
SET teachers = ARRAY(
        SELECT DISTINCT item->>'teacher'
        FROM jsonb_array_elements(teacher_auditoriums) AS item
        WHERE coalesce(item->>'teacher', '') <> ''
        ORDER BY 1
    ),
    auditoriums = ARRAY(
        SELECT DISTINCT item->>'auditorium'
        FROM jsonb_array_elements(teacher_auditoriums) AS item
        WHERE coalesce(item->>'auditorium', '') <> ''
        ORDER BY 1
    );

ALTER TABLE public.calendar_deleted_event
DROP COLUMN teacher_auditoriums;
