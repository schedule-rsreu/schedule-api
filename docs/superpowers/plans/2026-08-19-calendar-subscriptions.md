# Calendar Subscriptions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Добавить обновляемую iCalendar-подписку группы с изменениями через `SEQUENCE` и удалениями через `STATUS:CANCELLED`.

**Architecture:** PostgreSQL вычисляет стабильный UID и хранит общую ревизию плюс tombstone отменённых занятий. `schedule-parser2` обновляет tombstones в той же транзакции, где заменяет расписание. `schedule-api` объединяет активные и отменённые события и формирует `.ics` без новой runtime-зависимости.

**Tech Stack:** Go 1.26.5, Echo, sqlx, PostgreSQL 17, Python 3, psycopg2, goose v3.27.3 SQL migrations.

**Spec:** `docs/adr/18.08.2026-ice-design/18.08.2026-ice-design.md`

## Global Constraints

- Endpoint: `GET /api/v1/schedule/groups/{group}/calendar.ics` без дат в URL.
- Период: последние 14 дней и все доступные будущие занятия.
- Время событий хранится в БД как Moscow wall time и выдаётся в UTC с `Z`.
- Аудитории дедуплицируются и перечисляются через запятую; подгруппы не моделируются.
- Emoji включены глобально; пользовательской настройки пока нет.
- CI/CD и production-запуск goose не входят в эту реализацию.
- Все тесты запускаются нативно; Docker не используется.
- Не изменять незакоммиченные `tables.sql` и `gp.txt`.

---

### Task 1: Goose и календарная схема

**Files:**
- Create: `migrations/00001_calendar_subscriptions.sql`
- Modify: `Makefile`
- Modify: `README.md`

**Interfaces:**
- Produces: PostgreSQL functions/tables `calendar_event_uid`, `calendar_revision`, `calendar_deleted_event`.
- Produces: commands `make migrate-up`, `make migrate-down`, `make migrate-status` using `POSTGRES_DSN`.

- [ ] **Step 1: Add the SQL migration**

Migration `Up` creates a stable UID function, singleton revision row, and raw tombstone data:

```sql
-- +goose Up
CREATE FUNCTION calendar_event_uid(text, date, timestamp, text, text)
RETURNS text LANGUAGE sql STABLE AS $$
SELECT 'schedule-rsreu-' || md5(concat_ws('|', upper($1), $2::text, $3::time::text, $4, coalesce($5, ''))) || '@rsreu-schedule.ru'
$$;

CREATE TABLE calendar_revision (
    id smallint PRIMARY KEY CHECK (id = 1),
    revision bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO calendar_revision (id) VALUES (1);

CREATE TABLE calendar_deleted_event (
    uid text PRIMARY KEY,
    group_number text NOT NULL,
    start_time timestamp NOT NULL,
    end_time timestamp NOT NULL,
    title text NOT NULL,
    lesson_type text,
    teachers text[] NOT NULL DEFAULT '{}',
    auditoriums text[] NOT NULL DEFAULT '{}',
    sequence bigint NOT NULL,
    cancelled_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX calendar_deleted_event_group_start_idx
ON calendar_deleted_event (group_number, start_time);
```

`Down` drops the index, table, revision table, and function in reverse order.

- [ ] **Step 2: Add goose commands**

Extend `Makefile`:

```make
$(BIN)/goose:
	go install github.com/pressly/goose/v3/cmd/goose@v3.27.3

migrate-up: $(BIN)/goose
	$(BIN)/goose -dir migrations postgres "$(POSTGRES_DSN)" up

migrate-down: $(BIN)/goose
	$(BIN)/goose -dir migrations postgres "$(POSTGRES_DSN)" down

migrate-status: $(BIN)/goose
	$(BIN)/goose -dir migrations postgres "$(POSTGRES_DSN)" status
```

- [ ] **Step 3: Verify migration round-trip**

Run against an existing development PostgreSQL when `POSTGRES_DSN` is available:

```powershell
make migrate-up
make migrate-status
make migrate-down
make migrate-up
```

Expected: all commands exit `0`; final version is `1`.

- [ ] **Step 4: Document local migration commands**

Add the three commands and `POSTGRES_DSN` requirement to `README.md`.

- [ ] **Step 5: Commit**

```bash
git add migrations Makefile README.md
git commit -m "build: add goose database migrations"
```

---

### Task 2: Parser cancellation synchronization

**Files:**
- Modify: `D:/MyProjects/schedule-parser2/db.py`
- Create: `D:/MyProjects/schedule-parser2/test/test_calendar_sync.py`

**Interfaces:**
- Consumes: schema from Task 1.
- Produces: `begin_calendar_sync(cur) -> int` and `finish_calendar_sync(cur) -> None`.

- [ ] **Step 1: Write failing parser tests**

Tests use a minimal recording cursor and verify transaction orchestration:

```python
def test_begin_calendar_sync_returns_incremented_revision():
    cursor = RecordingCursor(fetchone=(7,))
    assert begin_calendar_sync(cursor) == 7
    assert "UPDATE calendar_revision" in cursor.statements[0]

def test_empty_schedule_is_rejected_before_delete():
    with pytest.raises(ValueError, match="empty schedule"):
        validate_schedule_data([])
```

- [ ] **Step 2: Run tests to verify RED**

Run: `python -m pytest test/test_calendar_sync.py -q`

Expected: FAIL because functions do not exist. If pytest is unavailable, use `python -m unittest` and equivalent `unittest.TestCase` assertions without adding a dependency.

- [ ] **Step 3: Implement minimal sync helpers**

`begin_calendar_sync` increments the revision and snapshots all future lessons using `calendar_event_uid`, `array_agg(DISTINCT teacher.full_name)` and `array_agg(DISTINCT auditorium.display_name)`. `finish_calendar_sync` removes tombstones whose UID exists in the newly inserted schedule and purges cancellations whose `end_time < now() - interval '30 days'`.

Call order inside the existing transaction:

```python
validate_schedule_data(data)
revision = begin_calendar_sync(cur)
delete_lessons_after_today(cur)
insert_new_schedule(cur, data)
finish_calendar_sync(cur)
```

- [ ] **Step 4: Run parser checks**

Run:

```powershell
python -m pytest test/test_calendar_sync.py -q
python -m compileall -q . -x "(\.venv|\.git|__pycache__|test/main\.py)"
```

Expected: both commands exit `0`.

- [ ] **Step 5: Commit in parser repository**

```bash
git add db.py test/test_calendar_sync.py
git commit -m "feat: track cancelled calendar events"
```

---

### Task 3: iCalendar formatter

**Files:**
- Create: `internal/models/calendar.go`
- Create: `internal/services/calendar.go`
- Create: `internal/services/calendar_test.go`

**Interfaces:**
- Produces: `models.CalendarEvent`, `models.GroupCalendar`.
- Produces: `GenerateCalendar(calendar *models.GroupCalendar) []byte`.

- [ ] **Step 1: Write failing formatter tests**

Cover one active multi-teacher/multi-room event and one cancelled event. Assert exact properties:

```go
assert.Contains(t, result, "SUMMARY:🧪 Проектирование информационных систем\r\n")
assert.Contains(t, result, "LOCATION:106а C, 110 C · РГРТУ\r\n")
assert.Contains(t, result, "STATUS:CANCELLED\r\n")
assert.Contains(t, result, "SEQUENCE:2\r\n")
assert.True(t, utf8.ValidString(result))
```

Add a line-folding case with Cyrillic longer than 75 octets.

- [ ] **Step 2: Run tests to verify RED**

Run: `go test ./internal/services -run Calendar -v`

Expected: FAIL because calendar models and formatter do not exist.

- [ ] **Step 3: Implement minimal formatter**

Use standard library only. Map types to `📘`, `✏️`, `🧪`, or `🎓`; sort/deduplicate teachers and auditoriums; convert Moscow wall time to UTC; escape text; fold without splitting UTF-8; emit CRLF, `GEO`, Google Maps `URL`, `STATUS`, `TRANSP`, and `SEQUENCE`.

- [ ] **Step 4: Run formatter tests to verify GREEN**

Run: `go test ./internal/services -run Calendar -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/models/calendar.go internal/services/calendar.go internal/services/calendar_test.go
git commit -m "feat: generate iCalendar feeds"
```

---

### Task 4: Repository and HTTP endpoint

**Files:**
- Modify: `internal/repo/repo.go`
- Modify: `internal/services/schedule.go`
- Modify: `internal/http/handlers/v1/schedule.go`
- Modify generated Swagger files under `docs/`.

**Interfaces:**
- Produces: `ScheduleRepo.GetGroupCalendar(ctx context.Context, group string) (*models.GroupCalendar, error)`.
- Produces: `ScheduleService.GetGroupCalendar(ctx context.Context, group string) ([]byte, error)`.
- Produces: `GET /api/v1/schedule/groups/:group/calendar.ics`.

- [ ] **Step 1: Write a failing service behavior test**

Add a repository query fixture test around the JSON shape returned for empty arrays, active events, and cancelled events. The production change that makes it pass is `GetGroupCalendar` returning a valid empty calendar for an existing group instead of `ErrNoResults`.

- [ ] **Step 2: Run the targeted test to verify RED**

Run: `go test ./internal/repo ./internal/services -run Calendar -v`

Expected: FAIL because repository/service methods do not exist.

- [ ] **Step 3: Implement repository and service**

Repository query joins group, lesson, teachers and auditoriums; computes active UID through `calendar_event_uid`; reads current `calendar_revision`; unions `calendar_deleted_event`; filters from `current_date - 14` through all future rows. Service uppercases group and calls `GenerateCalendar`.

- [ ] **Step 4: Implement handler**

Set:

```http
Content-Type: text/calendar; charset=utf-8
Content-Disposition: inline; filename="schedule-<group>.ics"
Cache-Control: public, max-age=300
```

Return `404` for unknown group and JSON errors for failures.

- [ ] **Step 5: Regenerate Swagger and run checks**

Run:

```powershell
make swag
go test ./...
go build ./cmd/main.go
```

Expected: all commands exit `0`.

- [ ] **Step 6: Commit**

```bash
git add internal docs
git commit -m "feat: expose group calendar subscriptions"
```

---

### Task 5: Cross-repository verification

**Files:**
- Modify ADR only if implementation reveals a mismatch.

- [ ] **Step 1: Apply migration to development database when available**

Run `make migrate-up` and confirm `make migrate-status` reports version `1`.

- [ ] **Step 2: Run parser once against development data**

Confirm revision increments, unchanged lessons remove false tombstones, and a deliberately removed lesson remains in `calendar_deleted_event`.

- [ ] **Step 3: Fetch and validate feed**

```powershell
curl.exe -f "http://localhost/api/v1/schedule/groups/344/calendar.ics" -o calendar.ics
```

Confirm active events, changed `SEQUENCE`, and deleted `STATUS:CANCELLED` entries are present.

- [ ] **Step 4: Run final repository checks**

Run API tests/build and parser tests/compileall again. Both repositories must have only the pre-existing untracked `tables.sql` and `gp.txt`.

- [ ] **Step 5: Record deployment follow-up**

Do not change CI/CD in this feature. Next work item: run `goose up` as a one-shot deployment job before rolling out parser/API containers.
