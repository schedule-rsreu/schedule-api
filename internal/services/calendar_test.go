package services_test

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"github.com/schedule-rsreu/schedule-api/internal/services"
)

func TestGenerateCalendar(t *testing.T) {
	calendar := &models.GroupCalendar{
		Group:     "344",
		UpdatedAt: time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC),
		Events: []models.CalendarEvent{
			{
				UID:         "active@rsreu-schedule.ru",
				StartTime:   time.Date(2026, 8, 19, 13, 35, 0, 0, time.UTC),
				EndTime:     time.Date(2026, 8, 19, 15, 10, 0, 0, time.UTC),
				Title:       "Проектирование информационных систем с очень длинным названием для проверки переноса строки",
				LessonType:  "lab",
				Teachers:    []string{"Соловьев Александр Вадимович", "Бурмистров Александр Сергеевич"},
				Auditoriums: []string{"110 C", "106а C", "110 C"},
				Sequence:    3,
			},
			{
				UID:        "cancelled@rsreu-schedule.ru",
				StartTime:  time.Date(2026, 8, 20, 15, 20, 0, 0, time.UTC),
				EndTime:    time.Date(2026, 8, 20, 16, 55, 0, 0, time.UTC),
				Title:      "Философия",
				LessonType: "lecture",
				Sequence:   4,
				Cancelled:  true,
			},
		},
	}

	result := string(services.GenerateCalendar(calendar))
	unfolded := strings.ReplaceAll(result, "\r\n ", "")

	for _, expected := range []string{
		"DTSTAMP:20260819T120000Z\r\n",
		"DTSTART:20260819T103500Z\r\n",
		"SUMMARY:🧪 Проектирование информационных систем",
		"LOCATION:106а C\\, 110 C · РГРТУ\r\n",
		"Преподаватели: Бурмистров Александр Сергеевич\\, Соловьев Александр Вадимович",
		"STATUS:CANCELLED\r\n",
		"SEQUENCE:4\r\n",
	} {
		if !strings.Contains(unfolded, expected) {
			t.Errorf("calendar does not contain %q:\n%s", expected, result)
		}
	}

	if !utf8.ValidString(result) {
		t.Fatal("calendar is not valid UTF-8")
	}
	for _, line := range strings.Split(result, "\r\n") {
		if len(line) > 75 {
			t.Fatalf("line is longer than 75 octets: %q", line)
		}
	}
}
