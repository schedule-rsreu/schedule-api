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

func TestGenerateCalendarLessonTypes(t *testing.T) {
	tests := []struct {
		lessonType string
		expected   string
	}{
		{"lecture", "📘 Предмет\\r\\nDESCRIPTION:Лекция"},
		{"lab", "🧪 Предмет\\r\\nDESCRIPTION:Лабораторная работа"},
		{"practice", "✏️ Предмет\\r\\nDESCRIPTION:Практика"},
		{"coursework", "📝 Предмет\\r\\nDESCRIPTION:Курсовая работа"},
		{"course_project", "📝 Предмет\\r\\nDESCRIPTION:Курсовой проект"},
		{"exam", "🎓 Предмет\\r\\nDESCRIPTION:Экзамен"},
		{"zachet", "✅ Предмет\\r\\nDESCRIPTION:Зачёт"},
		{"consultation", "💬 Предмет\\r\\nDESCRIPTION:Консультация"},
		{"elective", "⭐ Предмет\\r\\nDESCRIPTION:Факультатив"},
		{"unknown - другое", "🎓 Предмет\\r\\nDESCRIPTION:Занятие"},
	}

	for _, test := range tests {
		t.Run(test.lessonType, func(t *testing.T) {
			calendar := &models.GroupCalendar{
				Group:     "344",
				UpdatedAt: time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC),
				Events: []models.CalendarEvent{{
					UID:        test.lessonType + "@rsreu-schedule.ru",
					StartTime:  time.Date(2026, 8, 19, 13, 35, 0, 0, time.UTC),
					EndTime:    time.Date(2026, 8, 19, 15, 10, 0, 0, time.UTC),
					Title:      "Предмет",
					LessonType: test.lessonType,
				}},
			}

			result := strings.ReplaceAll(string(services.GenerateCalendar(calendar)), "\r\n ", "")
			expected := strings.ReplaceAll(test.expected, "\\r\\n", "\r\n")
			if !strings.Contains(result, "SUMMARY:"+expected) {
				t.Fatalf("unexpected lesson presentation:\n%s", result)
			}
		})
	}
}
