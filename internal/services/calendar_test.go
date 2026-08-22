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
		Source:    "https://api.example.com/api/v1/schedule/groups/344/calendar.ics",
		UpdatedAt: time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC),
		Events: []models.CalendarEvent{
			{
				UID:        "active@rsreu-schedule.ru",
				StartTime:  time.Date(2026, 8, 19, 13, 35, 0, 0, time.UTC),
				EndTime:    time.Date(2026, 8, 19, 15, 10, 0, 0, time.UTC),
				Title:      "Проектирование информационных систем с очень длинным названием для проверки переноса строки",
				LessonType: "lab",
				TeacherAuditoriums: []models.CalendarTeacherAuditorium{
					{Teacher: "Бурмистров Александр Сергеевич", Auditorium: "106а C"},
					{Teacher: "Соловьев Александр Вадимович", Auditorium: "110 C"},
				},
				Sequence: 3,
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
		"NAME:Расписание группы 344\r\n",
		"SOURCE;VALUE=URI:https://api.example.com/api/v1/schedule/groups/344/calendar.ics\r\n",
		"REFRESH-INTERVAL;VALUE=DURATION:PT1H\r\n",
		"COLOR:#5288c1\r\n",
		"DTSTAMP:20260819T120000Z\r\n",
		"DTSTART:20260819T103500Z\r\n",
		"SUMMARY:🧪 Лаб. Проектирование информационных систем с очень длинным названием для проверки переноса строки 106а C\\, 110 C\r\n",
		"CATEGORIES:EDUCATION,LAB\r\n",
		"LOCATION:106а C\\, 110 C · РГРТУ\r\n",
		"DESCRIPTION:Лабораторная работа\\nПроектирование информационных систем с очень длинным названием для проверки переноса строки\\nБурмистров Александр Сергеевич — 106а C\\nСоловьев Александр Вадимович — 110 C\\n\\nПерерыв: 14:20–14:25\r\n",
		"TRIGGER:-PT30M\r\n",
		"ACTION:DISPLAY\r\n",
		"STATUS:CANCELLED\r\n",
		"SEQUENCE:4\r\n",
	} {
		if !strings.Contains(unfolded, expected) {
			t.Errorf("calendar does not contain %q:\n%s", expected, result)
		}
	}
	if strings.Count(unfolded, "BEGIN:VALARM\r\n") != 1 {
		t.Fatalf("expected one alarm for the active event:\n%s", result)
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
		emoji      string
		name       string
		shortName  string
	}{
		{"lecture", "📘", "Лекция", "Лек."},
		{"lab", "🧪", "Лабораторная работа", "Лаб."},
		{"practice", "✏️", "Практика", "Упр."},
		{"coursework", "📄", "Курсовая работа", "Курс. раб."},
		{"course_project", "🛠️", "Курсовой проект", "Курс. пр."},
		{"exam", "🎓", "Экзамен", "Экз."},
		{"zachet", "🎓", "Зачёт", "Зач."},
		{"consultation", "❓", "Консультация", "Конс."},
		{"elective", "🧭", "Факультатив", "Фак."},
		{"unknown - другое", "🎓", "Занятие", "Зан."},
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
			if !strings.Contains(result, "SUMMARY:"+test.emoji+" "+test.shortName+" Предмет\r\n") ||
				!strings.Contains(result, "DESCRIPTION:"+test.name+"\\nПредмет\\n\\nПерерыв: 14:20–14:25\r\n") {
				t.Fatalf("unexpected lesson presentation:\n%s", result)
			}
		})
	}
}
