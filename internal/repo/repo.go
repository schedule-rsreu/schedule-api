package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"

	"github.com/schedule-rsreu/schedule-api/pkg/postgres"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

const queryTimeout = 10 * time.Second
const maxFacultyShortLen = 8
const maxDepartmentShortLen = 6

type ScheduleRepo struct {
	mdb                        *mongo.Database
	scheduleCollection         *mongo.Collection
	teachersScheduleCollection *mongo.Collection
	pg                         *postgres.Postgres
}

const ScheduleCollectionName = "schedule"
const TeachersScheduleCollectionName = "teachers_schedule"

func NewScheduleRepo(mdb *mongo.Database, pg *postgres.Postgres) *ScheduleRepo {
	scheduleCollection := mdb.Collection(ScheduleCollectionName)
	teachersScheduleCollection := mdb.Collection(TeachersScheduleCollectionName)
	return &ScheduleRepo{mdb, scheduleCollection, teachersScheduleCollection, pg}
}

func findOne[T any](filter any, c *mongo.Collection) (*T, error) {
	var result *T
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := c.FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNoResults
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func findOneJson[T any](pg *sqlx.DB, query string, args ...any) (*T, error) {
	var resultBytes []byte
	err := pg.QueryRow(query, args...).Scan(&resultBytes)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResults
		}
		return nil, fmt.Errorf("findOneJson: %w", err)
	}

	var result *T
	err = json.Unmarshal(resultBytes, &result)

	if err != nil {
		return nil, fmt.Errorf("findOneJson: json.Unmarshal: %w", err)
	}
	return result, err
}

func findAll[T any](filter any, c *mongo.Collection) ([]*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var results []*T

	cursor, err := c.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func aggregateOne[T any](pipeline any, c *mongo.Collection) (*T, error) {
	var results []*T

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	cursor.Next(context.TODO())
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		return result, nil
	}

	return nil, ErrNoResults
}

func aggregateAll[T any](pipeline any, c *mongo.Collection) ([]*T, error) {
	var results []*T

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrNoResults
	}

	return results, nil
}

func (sr *ScheduleRepo) groupExists(group string) (bool, error) {
	var exists bool
	err := sr.pg.DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM "group" WHERE number = $1)`, group).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (sr *ScheduleRepo) GetScheduleByGroup(group string, startDate, endDate time.Time) (*models.StudentSchedule, error) {
	const query = `
WITH params AS (
  SELECT
    $1::date AS start_date,
    $2::date AS end_date,
    $3::varchar AS group_number,
    g.id AS group_id,
    f.title_short AS faculty,
    g.course AS course
  FROM "group" g
  JOIN faculty f ON g.faculty_id = f.id
  WHERE g.number = $3
),

weekdays AS (
  SELECT unnest(ARRAY[
    'monday'::text,
    'tuesday',
    'wednesday',
    'thursday',
    'friday',
    'saturday'
  ]) AS weekday
),

lesson_rows AS (
  SELECT
    l.id AS lesson_id,
    trim(lower(to_char(l.date, 'Day'))) AS weekday,
    l.time,
    l.title,
    l.type,
    l.week_type,
    to_char(l.date, 'YYYY-MM-DD') AS date,
    l.start_time,
    l.end_time,
    lat.teacher_id,
    t.short_name,
    t.full_name,
    lat.auditorium_id,
    a.number AS auditorium_number,
	b.id AS building_id,
    b.letter AS building_letter,
    b.title AS building_title,
    l.date AS raw_date
  FROM lesson l
  JOIN params p ON l.group_id = p.group_id
  LEFT JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
  LEFT JOIN teacher t ON lat.teacher_id = t.id
  LEFT JOIN auditorium a ON lat.auditorium_id = a.id
  LEFT JOIN building b ON a.building_id = b.id
  WHERE l.date BETWEEN p.start_date AND p.end_date
),

lesson_core AS (
  SELECT DISTINCT
    lesson_id,
    weekday,
    time,
    title,
    type,
    week_type,
    date,
    start_time,
    end_time,
    raw_date
  FROM lesson_rows
),

teacher_aud_data AS (
  SELECT
    lesson_id,
    jsonb_build_object(
      'teacher', CASE
        WHEN teacher_id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', teacher_id,
          'short_name', short_name,
          'full_name', full_name
        )
      END,
      'auditorium', CASE
        WHEN auditorium_id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', auditorium_id,
          'number', auditorium_number,
          'display_name', auditorium_number || ' ' || building_letter,
          'building', jsonb_build_object(
			'id', building_id,
            'letter', building_letter,
            'title', building_title
          )
        )
      END
    ) AS teacher_auditorium
  FROM lesson_rows
),

lesson_with_teachers AS (
  SELECT
    l.*,
    jsonb_agg(t.teacher_auditorium ORDER BY t.teacher_auditorium) FILTER (WHERE t.teacher_auditorium IS NOT NULL) AS teacher_auditoriums
  FROM lesson_core l
  LEFT JOIN teacher_aud_data t ON l.lesson_id = t.lesson_id
  GROUP BY l.lesson_id, l.weekday, l.time, l.title, l.type, l.week_type, l.date, l.start_time, l.end_time, l.raw_date
),

-- Определяем точные недели из входных параметров
input_weeks AS (
  SELECT
    p.start_date AS first_week_monday,
    p.start_date + INTERVAL '7 days' AS second_week_monday,
    p.*
  FROM params p
),

-- Референсные занятия до/после для классификации
reference_lesson AS (
  SELECT
    iw.*,
    rl_before.lesson_date AS ref_lesson_date_before,
    rl_before.week_type AS ref_lesson_type_before,
    rl_after.lesson_date AS ref_lesson_date_after,
    rl_after.week_type AS ref_lesson_type_after
  FROM input_weeks iw
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    JOIN params p ON l.group_id = p.group_id
    WHERE l.date < iw.first_week_monday
    ORDER BY l.date DESC
    LIMIT 1
  ) rl_before ON true
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    JOIN params p ON l.group_id = p.group_id
    WHERE l.date > iw.second_week_monday + INTERVAL '6 days'
    ORDER BY l.date ASC
    LIMIT 1
  ) rl_after ON true
),

week_classification AS (
  SELECT
    rl.*,
    (
      SELECT l.week_type FROM lesson_with_teachers l 
      WHERE l.raw_date BETWEEN rl.first_week_monday 
        AND rl.first_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS first_week_actual_type,
    (
      SELECT l.week_type FROM lesson_with_teachers l 
      WHERE l.raw_date BETWEEN rl.second_week_monday 
        AND rl.second_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS second_week_actual_type,
    CASE 
      WHEN rl.ref_lesson_date_before IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_before = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      WHEN rl.ref_lesson_date_after IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_after = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      ELSE 'numerator'
    END AS predicted_first_week_type
  FROM reference_lesson rl
),

final_week_types AS (
  SELECT
    wc.*,
    COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) AS final_first_week_type,
    COALESCE(
      wc.second_week_actual_type, 
      CASE WHEN COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) = 'numerator' 
        THEN 'denominator' ELSE 'numerator' END
    ) AS final_second_week_type
  FROM week_classification wc
),

period_strings AS (
  SELECT
    CASE WHEN fwt.final_first_week_type = 'numerator' 
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS numerator_period,
    CASE WHEN fwt.final_first_week_type = 'denominator' 
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS denominator_period,
    fwt.final_first_week_type AS input_week_type
  FROM final_week_types fwt
),

grouped_lessons AS (
  SELECT
    week_type,
    weekday,
    json_agg(
      jsonb_build_object(
        'lesson', title,
        'type', type,
        'date', date,
        'time', time,
        'start_time', start_time,
        'end_time', end_time,
        'teacher_auditoriums', COALESCE(teacher_auditoriums, '[]'::jsonb)
      ) ORDER BY start_time
    ) AS lessons
  FROM lesson_with_teachers
  GROUP BY week_type, weekday
),

-- объединённые времена занятий по обеим неделям (для поля lessons_times)
lessons_times AS (
  SELECT array_agg(DISTINCT time ORDER BY time) AS lessons_times
  FROM lesson_with_teachers
  WHERE week_type IN (
    (SELECT input_week_type FROM period_strings),
    CASE WHEN (SELECT input_week_type FROM period_strings) = 'numerator' THEN 'denominator' ELSE 'numerator' END
  )
),

numerator_raw AS (
  SELECT weekday, lessons FROM grouped_lessons WHERE week_type = 'numerator'
),
denominator_raw AS (
  SELECT weekday, lessons FROM grouped_lessons WHERE week_type = 'denominator'
),

numerator_filled AS (
  SELECT w.weekday, COALESCE(n.lessons, '[]'::json) AS lessons
  FROM weekdays w
  LEFT JOIN numerator_raw n ON w.weekday = n.weekday
),
denominator_filled AS (
  SELECT w.weekday, COALESCE(d.lessons, '[]'::json) AS lessons
  FROM weekdays w
  LEFT JOIN denominator_raw d ON w.weekday = d.weekday
)

SELECT json_build_object(
  'faculty', p.faculty,
  'group', p.group_number,
  'course', p.course,
  'numerator_period', ps.numerator_period,
  'denominator_period', ps.denominator_period,
  'input_week_type', ps.input_week_type,
  'schedule', json_build_object(
    'numerator', json_build_object(
      'monday',    (SELECT lessons FROM numerator_filled WHERE weekday = 'monday'),
      'tuesday',   (SELECT lessons FROM numerator_filled WHERE weekday = 'tuesday'),
      'wednesday', (SELECT lessons FROM numerator_filled WHERE weekday = 'wednesday'),
      'thursday',  (SELECT lessons FROM numerator_filled WHERE weekday = 'thursday'),
      'friday',    (SELECT lessons FROM numerator_filled WHERE weekday = 'friday'),
      'saturday',  (SELECT lessons FROM numerator_filled WHERE weekday = 'saturday')
    ),
    'denominator', json_build_object(
      'monday',    (SELECT lessons FROM denominator_filled WHERE weekday = 'monday'),
      'tuesday',   (SELECT lessons FROM denominator_filled WHERE weekday = 'tuesday'),
      'wednesday', (SELECT lessons FROM denominator_filled WHERE weekday = 'wednesday'),
      'thursday',  (SELECT lessons FROM denominator_filled WHERE weekday = 'thursday'),
      'friday',    (SELECT lessons FROM denominator_filled WHERE weekday = 'friday'),
      'saturday',  (SELECT lessons FROM denominator_filled WHERE weekday = 'saturday')
    )
  ),
  'lessons_times', (SELECT lessons_times FROM lessons_times)
) AS schedule_json
FROM params p
CROSS JOIN period_strings ps;
`

	return findOneJson[models.StudentSchedule](sr.pg.DB, query, startDate, endDate, group)
}

// TODO: refactor to postgres
func (sr *ScheduleRepo) GetSchedulesByGroups(startDate, endDate time.Time, groups []string) ([]*models.StudentSchedule, error) {
	const query = `
-- Запрос для получения расписания групп из списка []*models.StudentSchedule
WITH params AS (
  SELECT
    $1::date AS start_date,
    $2::date AS end_date,
    $3::text[] AS group_numbers -- массив номеров групп
),

selected_groups AS (
  SELECT
    g.id AS group_id,
    g.number AS group_number,
    f.title_short AS faculty,
    g.course AS course
  FROM "group" g
  JOIN faculty f ON g.faculty_id = f.id
  CROSS JOIN params p
  WHERE g.number = ANY(p.group_numbers)
),

weekdays AS (
  SELECT unnest(ARRAY[
    'monday'::text,
    'tuesday',
    'wednesday',
    'thursday',
    'friday',
    'saturday'
  ]) AS weekday
),

lesson_rows AS (
  SELECT
    sg.group_id,
    sg.group_number,
    sg.faculty,
    sg.course,
    l.id AS lesson_id,
    trim(lower(to_char(l.date, 'Day'))) AS weekday,
    l.time,
    l.title,
    l.type,
    l.week_type,
    to_char(l.date, 'YYYY-MM-DD') AS date,
    l.start_time,
    l.end_time,
    lat.teacher_id,
    t.short_name,
    t.full_name,
    lat.auditorium_id,
    a.number AS auditorium_number,
	b.id AS building_id,
    b.letter AS building_letter,
    b.title AS building_title,
    l.date AS raw_date
  FROM selected_groups sg
  LEFT JOIN lesson l ON l.group_id = sg.group_id
  LEFT JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
  LEFT JOIN teacher t ON lat.teacher_id = t.id
  LEFT JOIN auditorium a ON lat.auditorium_id = a.id
  LEFT JOIN building b ON a.building_id = b.id
  CROSS JOIN params p
  WHERE l.date IS NULL OR l.date BETWEEN p.start_date AND p.end_date
),

lesson_core AS (
  SELECT DISTINCT
    lr.group_id,
    lr.group_number,
    lr.faculty,
    lr.course,
    lr.lesson_id,
    lr.weekday,
    lr.time,
    lr.title,
    lr.type,
    lr.week_type,
    lr.date,
    lr.start_time,
    lr.end_time,
    lr.raw_date
  FROM lesson_rows lr
  WHERE lr.lesson_id IS NOT NULL
),

teacher_aud_data AS (
  SELECT
    lr.lesson_id,
    jsonb_build_object(
      'teacher', CASE
        WHEN lr.teacher_id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', lr.teacher_id,
          'short_name', lr.short_name,
          'full_name', lr.full_name
        )
      END,
      'auditorium', CASE
        WHEN lr.auditorium_id IS NULL THEN NULL
        ELSE jsonb_build_object(
          'id', lr.auditorium_id,
          'number', lr.auditorium_number,
          'display_name', lr.auditorium_number || ' ' || lr.building_letter,
          'building', jsonb_build_object(
			'id', lr.building_id,
            'letter', lr.building_letter,
            'title', lr.building_title
          )
        )
      END
    ) AS teacher_auditorium
  FROM lesson_rows lr
  WHERE lr.teacher_id IS NOT NULL OR lr.auditorium_id IS NOT NULL
),

lesson_with_teachers AS (
  SELECT
    l.*,
    jsonb_agg(t.teacher_auditorium ORDER BY t.teacher_auditorium) FILTER (WHERE t.teacher_auditorium IS NOT NULL) AS teacher_auditoriums
  FROM lesson_core l
  LEFT JOIN teacher_aud_data t ON l.lesson_id = t.lesson_id
  GROUP BY l.group_id, l.group_number, l.faculty, l.course, l.lesson_id, l.weekday, l.time, l.title, l.type, l.week_type, l.date, l.start_time, l.end_time, l.raw_date
),

-- Определяем точные недели из входных параметров для каждой группы
input_weeks AS (
  SELECT
    sg.group_id,
    sg.group_number,
    sg.faculty,
    sg.course,
    p.start_date AS first_week_monday,
    p.start_date + INTERVAL '7 days' AS second_week_monday
  FROM selected_groups sg
  CROSS JOIN params p
),

-- Референсные занятия до/после для классификации для каждой группы
reference_lesson AS (
  SELECT
    iw.*,
    rl_before.lesson_date AS ref_lesson_date_before,
    rl_before.week_type AS ref_lesson_type_before,
    rl_after.lesson_date AS ref_lesson_date_after,
    rl_after.week_type AS ref_lesson_type_after
  FROM input_weeks iw
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    WHERE l.group_id = iw.group_id
      AND l.date < iw.first_week_monday
    ORDER BY l.date DESC
    LIMIT 1
  ) rl_before ON true
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    WHERE l.group_id = iw.group_id
      AND l.date > iw.second_week_monday + INTERVAL '6 days'
    ORDER BY l.date ASC
    LIMIT 1
  ) rl_after ON true
),

week_classification AS (
  SELECT
    rl.*,
    (
      SELECT l.week_type FROM lesson_with_teachers l 
      WHERE l.group_id = rl.group_id
        AND l.raw_date BETWEEN rl.first_week_monday 
        AND rl.first_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS first_week_actual_type,
    (
      SELECT l.week_type FROM lesson_with_teachers l 
      WHERE l.group_id = rl.group_id
        AND l.raw_date BETWEEN rl.second_week_monday 
        AND rl.second_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS second_week_actual_type,
    CASE 
      WHEN rl.ref_lesson_date_before IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_before = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      WHEN rl.ref_lesson_date_after IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_after = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      ELSE 'numerator'
    END AS predicted_first_week_type
  FROM reference_lesson rl
),

final_week_types AS (
  SELECT
    wc.*,
    COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) AS final_first_week_type,
    COALESCE(
      wc.second_week_actual_type, 
      CASE WHEN COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) = 'numerator' 
        THEN 'denominator' ELSE 'numerator' END
    ) AS final_second_week_type
  FROM week_classification wc
),

period_strings AS (
  SELECT
    fwt.group_id,
    fwt.group_number,
    fwt.faculty,
    fwt.course,
    CASE WHEN fwt.final_first_week_type = 'numerator' 
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS numerator_period,
    CASE WHEN fwt.final_first_week_type = 'denominator' 
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS denominator_period,
    fwt.final_first_week_type AS input_week_type
  FROM final_week_types fwt
),

-- Функция для форматирования типа занятия
type_formatter AS (
  SELECT 
    lwt.*,
    CASE lwt.type
      WHEN 'lecture' THEN 'Лек.'
      WHEN 'lab' THEN 'Лаб.'
      WHEN 'practice' THEN 'Упр.'
      WHEN 'coursework' THEN 'Курс. раб.'
      WHEN 'course_project' THEN 'Курс. проект'
      WHEN 'exam' THEN 'Экз.'
      WHEN 'zachet' THEN 'Зач.'
      WHEN 'consultation' THEN 'Конс.'
      WHEN 'elective' THEN 'Факультатив'
      WHEN 'unknown' THEN ''
      ELSE COALESCE(lwt.type, '')
    END AS formatted_type,
    -- Формирование строки с преподавателями и аудиториями
    (
      SELECT string_agg(
        CASE 
          WHEN ta->>'teacher' IS NOT NULL AND ta->>'auditorium' IS NOT NULL THEN
            (ta->'teacher'->>'short_name') || ' ' || (ta->'auditorium'->>'display_name')
          WHEN ta->>'teacher' IS NOT NULL AND ta->>'auditorium' IS NULL THEN
            (ta->'teacher'->>'short_name')
          WHEN ta->>'teacher' IS NULL AND ta->>'auditorium' IS NOT NULL THEN
            (ta->'auditorium'->>'display_name')
          ELSE ''
        END,
        E'\n'
        ORDER BY ta
      )
      FROM jsonb_array_elements(COALESCE(lwt.teacher_auditoriums, '[]'::jsonb)) AS ta
    ) AS teacher_auditorium_string
  FROM lesson_with_teachers lwt
),

grouped_lessons AS (
  SELECT
    tf.group_id,
    tf.week_type,
    tf.weekday,
    json_agg(
      jsonb_build_object(
        'lesson', CASE 
          WHEN tf.formatted_type != '' AND tf.teacher_auditorium_string IS NOT NULL AND tf.teacher_auditorium_string != '' THEN
            tf.formatted_type || ' ' || tf.title || E',\n' || tf.teacher_auditorium_string
          WHEN tf.formatted_type != '' AND (tf.teacher_auditorium_string IS NULL OR tf.teacher_auditorium_string = '') THEN
            tf.formatted_type || ' ' || tf.title
          WHEN (tf.formatted_type = '' OR tf.formatted_type IS NULL) AND tf.teacher_auditorium_string IS NOT NULL AND tf.teacher_auditorium_string != '' THEN
            tf.title || E',\n' || tf.teacher_auditorium_string
          ELSE
            tf.title
        END,
        'title', tf.title,
        'type', tf.type,
        'date', tf.date,
        'time', tf.time,
        'start_time', tf.start_time,
        'end_time', tf.end_time,
        'teacher_auditoriums', COALESCE(tf.teacher_auditoriums, '[]'::jsonb)
      ) ORDER BY tf.start_time
    ) AS lessons
  FROM type_formatter tf
  GROUP BY tf.group_id, tf.week_type, tf.weekday
),

-- объединённые времена занятий по обеим неделям для каждой группы
lessons_times AS (
  SELECT 
    lwt.group_id,
    array_agg(DISTINCT lwt.time ORDER BY lwt.time) AS lessons_times
  FROM lesson_with_teachers lwt
  JOIN period_strings ps ON lwt.group_id = ps.group_id
  WHERE lwt.week_type IN (
    ps.input_week_type,
    CASE WHEN ps.input_week_type = 'numerator' THEN 'denominator' ELSE 'numerator' END
  )
  GROUP BY lwt.group_id
),

numerator_raw AS (
  SELECT gl.group_id, gl.weekday, gl.lessons FROM grouped_lessons gl WHERE gl.week_type = 'numerator'
),
denominator_raw AS (
  SELECT gl.group_id, gl.weekday, gl.lessons FROM grouped_lessons gl WHERE gl.week_type = 'denominator'
),

numerator_filled AS (
  SELECT 
    sg.group_id,
    w.weekday, 
    COALESCE(n.lessons, '[]'::json) AS lessons
  FROM selected_groups sg
  CROSS JOIN weekdays w
  LEFT JOIN numerator_raw n ON sg.group_id = n.group_id AND w.weekday = n.weekday
),
denominator_filled AS (
  SELECT 
    sg.group_id,
    w.weekday, 
    COALESCE(d.lessons, '[]'::json) AS lessons
  FROM selected_groups sg
  CROSS JOIN weekdays w
  LEFT JOIN denominator_raw d ON sg.group_id = d.group_id AND w.weekday = d.weekday
)

SELECT json_agg(
  json_build_object(
    'faculty', ps.faculty,
    'group', ps.group_number,
    'course', ps.course,
    'numerator_period', ps.numerator_period,
    'denominator_period', ps.denominator_period,
    'input_week_type', ps.input_week_type,
    'schedule', json_build_object(
      'numerator', json_build_object(
        'monday',    (SELECT lessons FROM numerator_filled WHERE group_id = ps.group_id AND weekday = 'monday'),
        'tuesday',   (SELECT lessons FROM numerator_filled WHERE group_id = ps.group_id AND weekday = 'tuesday'),
        'wednesday', (SELECT lessons FROM numerator_filled WHERE group_id = ps.group_id AND weekday = 'wednesday'),
        'thursday',  (SELECT lessons FROM numerator_filled WHERE group_id = ps.group_id AND weekday = 'thursday'),
        'friday',    (SELECT lessons FROM numerator_filled WHERE group_id = ps.group_id AND weekday = 'friday'),
        'saturday',  (SELECT lessons FROM numerator_filled WHERE group_id = ps.group_id AND weekday = 'saturday')
      ),
      'denominator', json_build_object(
        'monday',    (SELECT lessons FROM denominator_filled WHERE group_id = ps.group_id AND weekday = 'monday'),
        'tuesday',   (SELECT lessons FROM denominator_filled WHERE group_id = ps.group_id AND weekday = 'tuesday'),
        'wednesday', (SELECT lessons FROM denominator_filled WHERE group_id = ps.group_id AND weekday = 'wednesday'),
        'thursday',  (SELECT lessons FROM denominator_filled WHERE group_id = ps.group_id AND weekday = 'thursday'),
        'friday',    (SELECT lessons FROM denominator_filled WHERE group_id = ps.group_id AND weekday = 'friday'),
        'saturday',  (SELECT lessons FROM denominator_filled WHERE group_id = ps.group_id AND weekday = 'saturday')
      )
    ),
    'lessons_times', COALESCE(lt.lessons_times, ARRAY[]::text[])
  ) ORDER BY ps.group_number
) AS schedules_json
FROM period_strings ps
LEFT JOIN lessons_times lt ON ps.group_id = lt.group_id;`
	fmt.Println(startDate, endDate, groups)
	res, err := findOneJson[[]*models.StudentSchedule](sr.pg.DB, query, startDate, endDate, groups)
	if err != nil {
		return nil, err
	}
	return *res, err
}

func (sr *ScheduleRepo) GetGroups(facultyName string, course int) (*models.CourseFacultyGroups, error) { //nolint:funlen,lll // too long queries
	const query = `
SELECT jsonb_build_object(
  'faculty', f.title_short,
  'course', g.course,
  'groups', COALESCE(
    jsonb_agg(g.number ORDER BY
        (CASE WHEN g.number ~ 'М$' THEN 1 ELSE 0 END), -- сначала без М, потом с М
        (regexp_replace(g.number, 'М$', '')::int)       -- сортировка по числовой части
    ) FILTER (WHERE g.number IS NOT NULL),
    '[]'::jsonb
  )
) AS result
FROM faculty f
LEFT JOIN "group" g ON g.faculty_id = f.id AND g.course = $2
WHERE f.title_short = $1
GROUP BY f.title_short, g.course;
`

	return findOneJson[models.CourseFacultyGroups](sr.pg.DB, query, facultyName, course)
}

func (sr *ScheduleRepo) GetFaculties() (*models.Faculties, error) {
	const query = `
	SELECT jsonb_build_object(
	  'faculties',
	  COALESCE(
		jsonb_agg(f.title_short ORDER BY f.title_short),
		'[]'::jsonb
	  )
	) AS result
	FROM faculty f;
`
	return findOneJson[models.Faculties](sr.pg.DB, query)
}

func (sr *ScheduleRepo) GetFacultyCourses(facultyName string) (*models.FacultyCourses, error) {
	const query = `
	SELECT jsonb_build_object(
	  'faculty', f.title_short,
	  'courses', COALESCE(
		  jsonb_agg(DISTINCT g.course ORDER BY g.course),
		  '[]'::jsonb
	  )
	) AS result
	FROM faculty f
	LEFT JOIN "group" g ON g.faculty_id = f.id
	WHERE f.title_short = $1
	GROUP BY f.title_short;
`
	var FacultyCoursesJSON []byte
	err := sr.pg.DB.QueryRow(query, facultyName).Scan(&FacultyCoursesJSON)

	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}

	var schedule *models.FacultyCourses
	err = json.Unmarshal(FacultyCoursesJSON, &schedule)

	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	return schedule, err
}

func (sr *ScheduleRepo) GetFacultiesWithCourses() (*models.FacultiesCourses, error) {
	const query = `
	SELECT jsonb_agg(fc ORDER BY fc->>'faculty') AS result
	FROM (
	  SELECT jsonb_build_object(
		'faculty', f.title_short,
		'courses',
		  COALESCE(
			(
			  SELECT jsonb_agg(course_num ORDER BY course_num)
			  FROM (
				SELECT DISTINCT g.course AS course_num
				FROM "group" g
				WHERE g.faculty_id = f.id
			  ) courses_sub
			),
			'[]'::jsonb
		  )
	  ) AS fc
	  FROM faculty f
	) t;
`
	return findOneJson[models.FacultiesCourses](sr.pg.DB, query)
}

func (sr *ScheduleRepo) GetCourseFaculties(course int) (*models.CourseFaculties, error) {
	fmt.Println("sdgfg")
	const query = `
	SELECT jsonb_build_object(
	  'course', $1::int,
	  'faculties',
	  COALESCE(
		(
		  SELECT jsonb_agg(title_short ORDER BY title_short)
		  FROM (
			SELECT DISTINCT f.title_short
			FROM faculty f
			JOIN "group" g ON g.faculty_id = f.id
			WHERE g.course = $1::int
		  ) sub
		),
		'[]'::jsonb
	  )
	) AS result;
`
	return findOneJson[models.CourseFaculties](sr.pg.DB, query, course)
}

func (sr *ScheduleRepo) GetTeacherSchedule(teacherID int, startDate, endDate time.Time) (*models.TeacherSchedule, error) {

	fmt.Println("startDate:", startDate, "endDate:", endDate)

	const query = `
WITH teacher_info AS (
  SELECT
    t.id,
    t.full_name,
    t.short_name,
    t.link,
    json_agg(DISTINCT jsonb_build_object(
      'id', d.id,
      'title', d.title,
      'title_short', d.title_short,
      'faculty', jsonb_build_object(
        'id', f.id,
        'title', f.title,
        'title_short', f.title_short
      )
    )) FILTER (WHERE d.id IS NOT NULL) AS departments
  FROM teacher t
  LEFT JOIN teacher_department td ON td.teacher_id = t.id
  LEFT JOIN department d ON d.id = td.department_id
  LEFT JOIN faculty f ON f.id = d.faculty_id
  WHERE t.id = $3
  GROUP BY t.id, t.full_name, t.short_name, t.link
),

params AS (
  SELECT
    $1::date AS start_date,
    $2::date AS end_date,
    $3::integer AS teacher_id,
    $1::date AS first_week_monday, 
    $1::date + INTERVAL '7 days' AS second_week_monday,  
    $1::date AS reference_date
),

lesson_rows AS (
  SELECT
    l.id AS lesson_id,
    trim(lower(to_char(l.date, 'Day'))) AS weekday,
    l.time,
    l.title,
    l.type,
    l.week_type,
    to_char(l.date, 'YYYY-MM-DD') AS date,
    a.id AS auditorium_id,
    a.number AS auditorium_number,
	b.id AS building_id,
    b.letter AS building_letter,
    b.title AS building_title,
    l.date AS raw_date
  FROM lesson l
  JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
  JOIN teacher t ON t.id = lat.teacher_id
  LEFT JOIN auditorium a ON a.id = lat.auditorium_id
  LEFT JOIN building b ON a.building_id = b.id
  WHERE t.id = $3
    AND l.date BETWEEN (SELECT start_date FROM params)
                   AND (SELECT end_date FROM params)
),

-- Отдельный CTE для получения всех групп по урокам преподавателя
lesson_groups AS (
  SELECT DISTINCT
    l.time,
    l.date,
    l.week_type,
    g.number AS group_number,
    g.course,
    f.title_short AS faculty_short
  FROM lesson l
  JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
  JOIN teacher t ON t.id = lat.teacher_id
  JOIN "group" g ON g.id = l.group_id
  JOIN faculty f ON f.id = g.faculty_id
  WHERE t.id = $3
    AND l.date BETWEEN (SELECT start_date FROM params)
                   AND (SELECT end_date FROM params)
),

-- Получаем первый урок для каждой комбинации время+дата+преподаватель
lesson_representative AS (
  SELECT DISTINCT ON (l.time, l.date)
    l.id AS lesson_id,
    l.time,
    l.date,
    l.title,
    l.type,
    l.week_type
  FROM lesson l
  JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
  JOIN teacher t ON t.id = lat.teacher_id
  WHERE t.id = $3
    AND l.date BETWEEN (SELECT start_date FROM params)
                   AND (SELECT end_date FROM params)
  ORDER BY l.time, l.date, l.id
),

lesson_core AS (
  SELECT DISTINCT
    lesson_id,
    trim(lower(to_char(date, 'Day'))) AS weekday,
    time,
    title,
    type,
    week_type,
    to_char(date, 'YYYY-MM-DD') AS date,
    date AS raw_date
  FROM lesson_representative
),

lesson_with_auds AS (
  SELECT
    lc.*,
    (
      SELECT jsonb_agg(aud ORDER BY aud->>'number', aud->'building'->>'letter')
      FROM (
        SELECT DISTINCT jsonb_build_object(
          'id', lr.auditorium_id,
          'building', jsonb_build_object('id', lr.building_id, 'letter', lr.building_letter, 'title', lr.building_title),
          'number', lr.auditorium_number,
          'display_name', lr.auditorium_number || ' ' || lr.building_letter
        ) AS aud
        FROM lesson_rows lr
        WHERE lr.time = lc.time
          AND to_char(lr.raw_date, 'YYYY-MM-DD') = lc.date
          AND lr.auditorium_number IS NOT NULL
      ) sub
    ) AS auditoriums
  FROM lesson_core lc
),

reference_lesson AS (
  SELECT
    p.*,
    rl_before.lesson_date AS ref_lesson_date_before,
    rl_before.week_type AS ref_lesson_type_before,
    rl_after.lesson_date AS ref_lesson_date_after,
    rl_after.week_type AS ref_lesson_type_after
  FROM params p
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
    WHERE l.date < p.first_week_monday
      AND EXISTS (
        SELECT 1 FROM lesson_auditorium_teacher lat2
        JOIN teacher t2 ON t2.id = lat2.teacher_id
        WHERE lat2.lesson_id = l.id AND t2.id = $3
      )
    ORDER BY l.date DESC
    LIMIT 1
  ) rl_before ON true
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id
    WHERE l.date > p.second_week_monday + INTERVAL '6 days'
      AND EXISTS (
        SELECT 1 FROM lesson_auditorium_teacher lat2
        JOIN teacher t2 ON t2.id = lat2.teacher_id
        WHERE lat2.lesson_id = l.id AND t2.id = $3
      )
    ORDER BY l.date ASC
    LIMIT 1
  ) rl_after ON true
),

week_classification AS (
  SELECT
    rl.*,
    (
      SELECT l.week_type FROM lesson_with_auds l
      WHERE l.raw_date BETWEEN rl.first_week_monday AND rl.first_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS first_week_actual_type,
    (
      SELECT l.week_type FROM lesson_with_auds l
      WHERE l.raw_date BETWEEN rl.second_week_monday AND rl.second_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS second_week_actual_type,
    CASE
      WHEN rl.ref_lesson_date_before IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_before = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7*24*3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7*24*3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      WHEN rl.ref_lesson_date_after IS NOT NULL THEN
        CASE
          WHEN rl.ref_lesson_type_after = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7*24*3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7*24*3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      ELSE 'numerator'
    END AS predicted_first_week_type
  FROM reference_lesson rl
),

final_week_types AS (
  SELECT
    wc.*,
    COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) AS final_first_week_type,
    COALESCE(
      wc.second_week_actual_type,
      CASE WHEN COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) = 'numerator'
        THEN 'denominator' ELSE 'numerator' END
    ) AS final_second_week_type
  FROM week_classification wc
),

period_strings AS (
  SELECT
    fwt.*,
    -- Период числителя - всегда первая неделя, если она числитель, иначе вторая
    CASE WHEN fwt.final_first_week_type = 'numerator'
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS numerator_period,
    -- Период знаменателя - всегда первая неделя, если она знаменатель, иначе вторая
    CASE WHEN fwt.final_first_week_type = 'denominator'
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS denominator_period,
    -- Тип недели для входной даты
    CASE 
      WHEN fwt.reference_date BETWEEN fwt.first_week_monday AND fwt.first_week_monday + INTERVAL '6 days'
        THEN fwt.final_first_week_type
      ELSE fwt.final_second_week_type
    END AS input_week_type
  FROM final_week_types fwt
),

grouped_lessons AS (
  SELECT
    week_type,
    weekday,
    json_agg(
      jsonb_build_object(
        'time', time,
        'lesson', title,
        'type', type,
        'date', date,
        'faculties', (
          SELECT array_agg(DISTINCT lg.faculty_short ORDER BY lg.faculty_short)
          FROM lesson_groups lg
          WHERE lg.time = lesson_with_auds.time
            AND to_char(lg.date, 'YYYY-MM-DD') = lesson_with_auds.date
            AND lg.week_type = lesson_with_auds.week_type
            AND lg.faculty_short IS NOT NULL
        ),
        'groups', (
          SELECT array_agg(DISTINCT lg.group_number ORDER BY lg.group_number)
          FROM lesson_groups lg
          WHERE lg.time = lesson_with_auds.time
            AND to_char(lg.date, 'YYYY-MM-DD') = lesson_with_auds.date
            AND lg.week_type = lesson_with_auds.week_type
            AND lg.group_number IS NOT NULL
        ),
        'courses', (
          SELECT array_agg(DISTINCT lg.course ORDER BY lg.course)
          FROM lesson_groups lg
          WHERE lg.time = lesson_with_auds.time
            AND to_char(lg.date, 'YYYY-MM-DD') = lesson_with_auds.date
            AND lg.week_type = lesson_with_auds.week_type
            AND lg.course IS NOT NULL
        ),
        'auditorium', (
          CASE
            WHEN auditoriums IS NULL OR jsonb_array_length(auditoriums) = 0 THEN NULL
            ELSE (auditoriums->0)
          END
        )
      ) ORDER BY time
    ) AS lessons
  FROM lesson_with_auds
  GROUP BY week_type, weekday
),

numerator_raw AS (
  SELECT weekday, lessons FROM grouped_lessons WHERE week_type = 'numerator'
),
denominator_raw AS (
  SELECT weekday, lessons FROM grouped_lessons WHERE week_type = 'denominator'
),

numerator_filled AS (
  SELECT w.weekday, COALESCE(n.lessons, '[]'::json) AS lessons
  FROM (SELECT * FROM (VALUES ('monday'),('tuesday'),('wednesday'),('thursday'),('friday'),('saturday')) AS v(weekday)) w
  LEFT JOIN numerator_raw n ON w.weekday = n.weekday
),
denominator_filled AS (
  SELECT w.weekday, COALESCE(d.lessons, '[]'::json) AS lessons
  FROM (SELECT * FROM (VALUES ('monday'),('tuesday'),('wednesday'),('thursday'),('friday'),('saturday')) AS v(weekday)) w
  LEFT JOIN denominator_raw d ON w.weekday = d.weekday
),

lessons_times AS (
  SELECT array_agg(DISTINCT time ORDER BY time) AS lessons_times
  FROM lesson_with_auds
  WHERE week_type IN ('numerator', 'denominator')
)

SELECT json_build_object(
  'id', ti.id,
  'full_name', ti.full_name,
  'short_name', ti.short_name,
  'link', ti.link,
  'departments', ti.departments,
  'numerator_period', ps.numerator_period,
  'denominator_period', ps.denominator_period,
  'input_week_type', ps.input_week_type,
  'schedule', json_build_object(
    'numerator', json_build_object(
      'monday',    (SELECT lessons FROM numerator_filled WHERE weekday = 'monday'),
      'tuesday',   (SELECT lessons FROM numerator_filled WHERE weekday = 'tuesday'),
      'wednesday', (SELECT lessons FROM numerator_filled WHERE weekday = 'wednesday'),
      'thursday',  (SELECT lessons FROM numerator_filled WHERE weekday = 'thursday'),
      'friday',    (SELECT lessons FROM numerator_filled WHERE weekday = 'friday'),
      'saturday',  (SELECT lessons FROM numerator_filled WHERE weekday = 'saturday')
    ),
    'denominator', json_build_object(
      'monday',    (SELECT lessons FROM denominator_filled WHERE weekday = 'monday'),
      'tuesday',   (SELECT lessons FROM denominator_filled WHERE weekday = 'tuesday'),
      'wednesday', (SELECT lessons FROM denominator_filled WHERE weekday = 'wednesday'),
      'thursday',  (SELECT lessons FROM denominator_filled WHERE weekday = 'thursday'),
      'friday',    (SELECT lessons FROM denominator_filled WHERE weekday = 'friday'),
      'saturday',  (SELECT lessons FROM denominator_filled WHERE weekday = 'saturday')
    )
  ),
  'lessons_times', (SELECT lessons_times FROM lessons_times)
) AS schedule_json
FROM period_strings ps
JOIN teacher_info ti ON ti.id = $3;`

	return findOneJson[models.TeacherSchedule](sr.pg.DB, query, startDate, endDate, teacherID)
}

func (sr *ScheduleRepo) GetAllTeachers() (*models.TeachersList, error) {
	const query = `
SELECT json_build_object(
    'teachers', json_agg(
        json_build_object(
            'id', t.id,
            'full_name', t.full_name,
            'short_name', t.short_name
        )
    )
) AS teachers_list
FROM (
    SELECT 
        id,
        full_name,
        short_name
    FROM teacher
    ORDER BY full_name
) t;
`
	return findOneJson[models.TeachersList](sr.pg.DB, query)
}

func (sr *ScheduleRepo) GetTeachersFaculties(departmentID int) ([]*models.Faculty, error) {
	const query = `
-- Запрос для получения списка факультетов с фильтрацией по departmentID
SELECT json_agg(
    json_build_object(
        'id', f.id,
        'title', f.title,
        'title_short', f.title_short
    )
) AS faculties_list
FROM (
    SELECT DISTINCT
        f.id,
        f.title,
        f.title_short
    FROM faculty f
    WHERE CASE 
        WHEN $1 = 0 THEN true  -- если departmentID = 0, показать все факультеты
        ELSE EXISTS (
            SELECT 1 
            FROM department d 
            WHERE d.faculty_id = f.id 
            AND d.id = $1  -- фильтр по departmentID
        )
    END
    ORDER BY f.title
) f;`
	res, err := findOneJson[[]*models.Faculty](sr.pg.DB, query, departmentID)
	return *res, err
}

func (sr *ScheduleRepo) GetTeachersDepartments(facultyID int) ([]*models.Department, error) {
	const query = `
-- Запрос для получения списка кафедр с фильтрацией по faculty_id
SELECT json_agg(
    json_build_object(
        'id', d.id,
        'title', d.title,
        'title_short', d.title_short,
        'faculty', json_build_object(
            'id', f.id,
            'title', f.title,
            'title_short', f.title_short
        )
    )
) AS departments_list
FROM (
    SELECT 
        d.id,
        d.title,
        d.title_short,
        d.faculty_id,
        f.id as faculty_id_ref,
        f.title as faculty_title,
        f.title_short as faculty_title_short
    FROM department d
    JOIN faculty f ON d.faculty_id = f.id
    WHERE CASE 
        WHEN $1 = 0 THEN true  -- если faculty_id = 0, показать все кафедры
        ELSE d.faculty_id = $1  -- фильтр по faculty_id
    END
    ORDER BY d.title
) d
JOIN faculty f ON d.faculty_id = f.id;`
	res, err := findOneJson[[]*models.Department](sr.pg.DB, query, facultyID)
	return *res, err
}

func (sr *ScheduleRepo) GetTeachersList(facultyID, departmentID int) (*models.TeachersList, error) {
	const query = `
-- Запрос для получения списка преподавателей с фильтрацией по facultyID и departmentID
SELECT json_build_object(
    'teachers', json_agg(
        json_build_object(
            'id', teacher_data.id,
            'full_name', teacher_data.full_name,
            'short_name', teacher_data.short_name
        )
    )
) AS teachers_list
FROM (
    SELECT DISTINCT
        t.id,
        t.full_name,
        t.short_name
    FROM teacher t
    JOIN teacher_department td ON t.id = td.teacher_id
    JOIN department d ON td.department_id = d.id
    JOIN faculty f ON d.faculty_id = f.id
    WHERE 1=1
        AND CASE 
            WHEN $1 = 0 THEN true  -- если facultyID = 0, не фильтровать по факультету
            ELSE f.id = $1  -- фильтр по facultyID
        END
        AND CASE 
            WHEN $2 = 0 THEN true  -- если departmentID = 0, не фильтровать по кафедре
            ELSE d.id = $2  -- фильтр по departmentID
        END
    ORDER BY t.full_name
) teacher_data;
`
	return findOneJson[models.TeachersList](sr.pg.DB, query, facultyID, departmentID)
}

func (sr *ScheduleRepo) GetAuditoriumSchedule(startDate, endDate time.Time, auditoriumID int) (*models.AuditoriumSchedule, error) {
	const query = `
WITH params AS (
  SELECT
    $1::date AS start_date,
    $2::date AS end_date,
    $3::int AS auditorium_id
),

weekdays AS (
  SELECT unnest(ARRAY[
    'monday'::text,
    'tuesday',
    'wednesday',
    'thursday',
    'friday',
    'saturday'
  ]) AS weekday
),

lesson_rows AS (
  SELECT
    l.id AS lesson_id,
    trim(lower(to_char(l.date, 'Day'))) AS weekday,
    l.time,
    l.title,
    l.type,
    l.week_type,
    to_char(l.date, 'YYYY-MM-DD') AS date,
    l.start_time,
    l.end_time,
    l.group_id,
    g.number AS group_number,
    g.course,
    f.title_short AS faculty,
    lat.teacher_id,
    t.short_name,
    t.full_name,
    l.date AS raw_date
  FROM lesson l
  JOIN params p ON true
  JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id AND lat.auditorium_id = p.auditorium_id
  JOIN "group" g ON l.group_id = g.id
  JOIN faculty f ON g.faculty_id = f.id
  LEFT JOIN teacher t ON lat.teacher_id = t.id
  WHERE l.date BETWEEN p.start_date AND p.end_date
),

lesson_core AS (
  SELECT DISTINCT
    lesson_id,
    weekday,
    time,
    title,
    type,
    week_type,
    date,
    start_time,
    end_time,
    raw_date
  FROM lesson_rows
),

auditorium_info AS (
  SELECT 
    a.id,
    a.number,
    a.number || ' ' || b.letter AS display_name,
	b.id AS building_id,
    b.letter AS building_letter,
    b.title AS building_title
  FROM auditorium a
  JOIN building b ON a.building_id = b.id
  JOIN params p ON a.id = p.auditorium_id
),

-- Агрегируем данные о группах, факультетах, курсах и преподавателях для каждой комбинации время+дата+тип+название
lesson_aggregated_data AS (
  SELECT
    lr.weekday,
    lr.time,
    lr.title,
    lr.type,
    lr.week_type,
    lr.date,
    lr.start_time,
    lr.end_time,
    lr.raw_date,
    array_agg(DISTINCT lr.faculty ORDER BY lr.faculty) AS faculties,
    array_agg(DISTINCT lr.group_number ORDER BY lr.group_number) AS groups,
    array_agg(DISTINCT lr.course ORDER BY lr.course) AS courses,
    jsonb_agg(
      DISTINCT 
      CASE 
        WHEN lr.teacher_id IS NOT NULL THEN
          jsonb_build_object(
            'id', lr.teacher_id,
            'short_name', lr.short_name,
            'full_name', lr.full_name
          )
      END
    ) FILTER (WHERE lr.teacher_id IS NOT NULL) AS teachers
  FROM lesson_rows lr
  GROUP BY lr.weekday, lr.time, lr.title, lr.type, lr.week_type, lr.date, lr.start_time, lr.end_time, lr.raw_date
),

lesson_with_details AS (
  SELECT
    weekday,
    time,
    title,
    type,
    week_type,
    date,
    start_time,
    end_time,
    raw_date,
    faculties,
    groups,
    courses,
    COALESCE(teachers, '[]'::jsonb) AS teachers
  FROM lesson_aggregated_data
),

-- Определяем точные недели из входных параметров
input_weeks AS (
  SELECT
    p.start_date AS first_week_monday,
    p.start_date + INTERVAL '7 days' AS second_week_monday,
    p.*
  FROM params p
),

-- Референсные занятия до/после для классификации
reference_lesson AS (
  SELECT
    iw.*,
    rl_before.lesson_date AS ref_lesson_date_before,
    rl_before.week_type AS ref_lesson_type_before,
    rl_after.lesson_date AS ref_lesson_date_after,
    rl_after.week_type AS ref_lesson_type_after
  FROM input_weeks iw
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id AND lat.auditorium_id = iw.auditorium_id
    WHERE l.date < iw.first_week_monday
    ORDER BY l.date DESC
    LIMIT 1
  ) rl_before ON true
  LEFT JOIN LATERAL (
    SELECT l.date AS lesson_date, l.week_type
    FROM lesson l
    JOIN lesson_auditorium_teacher lat ON lat.lesson_id = l.id AND lat.auditorium_id = iw.auditorium_id
    WHERE l.date > iw.second_week_monday + INTERVAL '6 days'
    ORDER BY l.date ASC
    LIMIT 1
  ) rl_after ON true
),

week_classification AS (
  SELECT
    rl.*,
    (
      SELECT l.week_type FROM lesson_with_details l 
      WHERE l.raw_date BETWEEN rl.first_week_monday 
        AND rl.first_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS first_week_actual_type,
    (
      SELECT l.week_type FROM lesson_with_details l 
      WHERE l.raw_date BETWEEN rl.second_week_monday 
        AND rl.second_week_monday + INTERVAL '6 days'
      LIMIT 1
    ) AS second_week_actual_type,
    CASE 
      WHEN rl.ref_lesson_date_before IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_before = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (rl.first_week_monday::timestamp - date_trunc('week', rl.ref_lesson_date_before)::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      WHEN rl.ref_lesson_date_after IS NOT NULL THEN
        CASE 
          WHEN rl.ref_lesson_type_after = 'numerator' THEN
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'numerator' ELSE 'denominator' END
          ELSE
            CASE WHEN (EXTRACT(EPOCH FROM (date_trunc('week', rl.ref_lesson_date_after)::timestamp - rl.first_week_monday::timestamp)) / (7 * 24 * 3600))::int % 2 = 0
              THEN 'denominator' ELSE 'numerator' END
        END
      ELSE 'numerator'
    END AS predicted_first_week_type
  FROM reference_lesson rl
),

final_week_types AS (
  SELECT
    wc.*,
    COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) AS final_first_week_type,
    COALESCE(
      wc.second_week_actual_type, 
      CASE WHEN COALESCE(wc.first_week_actual_type, wc.predicted_first_week_type) = 'numerator' 
        THEN 'denominator' ELSE 'numerator' END
    ) AS final_second_week_type
  FROM week_classification wc
),

period_strings AS (
  SELECT
    CASE WHEN fwt.final_first_week_type = 'numerator' 
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS numerator_period,
    CASE WHEN fwt.final_first_week_type = 'denominator' 
      THEN to_char(fwt.first_week_monday, 'DD.MM') || '-' || to_char(fwt.first_week_monday + INTERVAL '6 days', 'DD.MM')
      ELSE to_char(fwt.second_week_monday, 'DD.MM') || '-' || to_char(fwt.second_week_monday + INTERVAL '6 days', 'DD.MM')
    END AS denominator_period,
    fwt.final_first_week_type AS input_week_type
  FROM final_week_types fwt
),

grouped_lessons AS (
  SELECT
    week_type,
    weekday,
    json_agg(
      jsonb_build_object(
        'time', time,
        'date', date,
        'type', type,
        'lesson', title,
        'faculties', faculties,
        'groups', groups,
        'courses', courses,
        'teachers', teachers
      ) ORDER BY start_time
    ) AS lessons
  FROM lesson_with_details
  GROUP BY week_type, weekday
),

numerator_raw AS (
  SELECT weekday, lessons FROM grouped_lessons WHERE week_type = 'numerator'
),
denominator_raw AS (
  SELECT weekday, lessons FROM grouped_lessons WHERE week_type = 'denominator'
),

numerator_filled AS (
  SELECT w.weekday, COALESCE(n.lessons, '[]'::json) AS lessons
  FROM weekdays w
  LEFT JOIN numerator_raw n ON w.weekday = n.weekday
),
denominator_filled AS (
  SELECT w.weekday, COALESCE(d.lessons, '[]'::json) AS lessons
  FROM weekdays w
  LEFT JOIN denominator_raw d ON w.weekday = d.weekday
)

SELECT json_build_object(
  'auditorium', json_build_object(
    'id', ai.id,
    'number', ai.number,
    'display_name', ai.display_name,
    'building', json_build_object(
	  'id', ai.building_id,
      'letter', ai.building_letter,
      'title', ai.building_title
    )
  ),
  'numerator_period', ps.numerator_period,
  'denominator_period', ps.denominator_period,
  'input_week_type', ps.input_week_type,
  'schedule', json_build_object(
    'numerator', json_build_object(
      'monday',    (SELECT lessons FROM numerator_filled WHERE weekday = 'monday'),
      'tuesday',   (SELECT lessons FROM numerator_filled WHERE weekday = 'tuesday'),
      'wednesday', (SELECT lessons FROM numerator_filled WHERE weekday = 'wednesday'),
      'thursday',  (SELECT lessons FROM numerator_filled WHERE weekday = 'thursday'),
      'friday',    (SELECT lessons FROM numerator_filled WHERE weekday = 'friday'),
      'saturday',  (SELECT lessons FROM numerator_filled WHERE weekday = 'saturday')
    ),
    'denominator', json_build_object(
      'monday',    (SELECT lessons FROM denominator_filled WHERE weekday = 'monday'),
      'tuesday',   (SELECT lessons FROM denominator_filled WHERE weekday = 'tuesday'),
      'wednesday', (SELECT lessons FROM denominator_filled WHERE weekday = 'wednesday'),
      'thursday',  (SELECT lessons FROM denominator_filled WHERE weekday = 'thursday'),
      'friday',    (SELECT lessons FROM denominator_filled WHERE weekday = 'friday'),
      'saturday',  (SELECT lessons FROM denominator_filled WHERE weekday = 'saturday')
    )
  )
) AS schedule_json
FROM auditorium_info ai
CROSS JOIN period_strings ps;
`
	return findOneJson[models.AuditoriumSchedule](sr.pg.DB, query, startDate, endDate, auditoriumID)
}

func (r *ScheduleRepo) GetAuditorium(auditoriumID int) (*models.Auditorium, error) {
	const query = `
        SELECT json_build_object(
            'id', a.id,
            'number', a.number,
            'display_name', a.number || ' ' || b.letter,
            'building', json_build_object(
                'id', b.id,
                'letter', b.letter,
                'title', b.title
            )
        ) AS auditorium_json
        FROM auditorium a
        JOIN building b ON a.building_id = b.id
        WHERE a.id = $1
    `
	return findOneJson[models.Auditorium](r.pg.DB, query, auditoriumID)
}

func (r *ScheduleRepo) GetAuditoriumsList(buildingId int) ([]*models.Auditorium, error) {
	const query = `
        SELECT json_agg(
            json_build_object(
                'id', a.id,
                'number', a.number,
                'display_name', a.number || ' ' || b.letter,
                'building', json_build_object(
                    'id', b.id,
                    'letter', b.letter,
                    'title', b.title
                )
            ) ORDER BY b.id, a.number
        ) AS auditoriums_json
        FROM auditorium a
        JOIN building b ON a.building_id = b.id
        WHERE ($1 = 0 OR b.id = $1)
    `
	res, err := findOneJson[[]*models.Auditorium](r.pg.DB, query, buildingId)
	return *res, err
}

func (r *ScheduleRepo) GetBuildingsList() ([]*models.Building, error) {
	const query = `
        SELECT json_agg(
            json_build_object(
                'id', b.id,
                'letter', b.letter,
                'title', b.title
            ) ORDER BY b.id
        ) AS buildings_json
        FROM building b
    `
	res, err := findOneJson[[]*models.Building](r.pg.DB, query)
	return *res, err
}

func (r *ScheduleRepo) GetBuilding(buildingId int) (*models.Building, error) {
	const query = `
        SELECT json_build_object(
            'id', b.id,
            'letter', b.letter,
            'title', b.title
        ) AS building_json
        FROM building b
        WHERE b.id = $1
    `
	return findOneJson[models.Building](r.pg.DB, query, buildingId)
}
