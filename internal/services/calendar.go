package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"github.com/schedule-rsreu/schedule-api/internal/repo"
)

const (
	calendarGeo = "54.6132708;39.7236472"
	calendarURL = "https://www.google.com/maps?q=54.6132708,39.7236472"
)

func (s *ScheduleService) GetGroupCalendar(ctx context.Context, group string, includeMilitary bool) ([]byte, error) {
	group = strings.ToUpper(strings.TrimSpace(group))
	calendar, err := s.Repo.GetGroupCalendar(ctx, group)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("calendar for group %v not found", group)}
		}
		return nil, err
	}
	if !includeMilitary {
		calendar.Events = slices.DeleteFunc(calendar.Events, func(event models.CalendarEvent) bool {
			if normalizeLessonValue(event.Title) != "военная подготовка" {
				return false
			}
			for _, pair := range event.TeacherAuditoriums {
				if isMilitaryAuditorium(pair.Auditorium) {
					return true
				}
			}
			return false
		})
	}
	return GenerateCalendar(calendar), nil
}

func GenerateCalendar(calendar *models.GroupCalendar) []byte {
	var result strings.Builder
	writeCalendarLine(&result, "BEGIN:VCALENDAR")
	writeCalendarLine(&result, "VERSION:2.0")
	writeCalendarLine(&result, "PRODID:-//schedule-rsreu//Schedule API//RU")
	writeCalendarLine(&result, "CALSCALE:GREGORIAN")
	writeCalendarLine(&result, "METHOD:PUBLISH")
	calendarName := "Расписание группы " + calendar.Group
	writeProperty(&result, "NAME", calendarName)
	writeProperty(&result, "X-WR-CALNAME", calendarName)
	writeCalendarLine(&result, "X-WR-TIMEZONE:Europe/Moscow")
	writeCalendarLine(&result, "REFRESH-INTERVAL;VALUE=DURATION:PT1H")
	writeCalendarLine(&result, "COLOR:#5288c1")

	dtstamp := calendar.UpdatedAt.UTC().Format("20060102T150405Z")
	for index := range calendar.Events {
		event := &calendar.Events[index]
		if event.Cancelled {
			continue
		}
		emoji, lessonTypeName, lessonTypeShortName := lessonPresentation(event.LessonType)
		auditoriums := eventAuditoriums(event)
		summary := emoji + " " + lessonTypeShortName + " " + event.Title
		if len(auditoriums) > 0 {
			summary += " " + strings.Join(auditoriums, ", ")
		}
		alarmDescription := lessonTypeShortName + " " + event.Title
		if len(auditoriums) > 0 {
			alarmDescription += " " + strings.Join(auditoriums, ", ")
		}
		alarmDescription += " через 20 минут"
		writeCalendarLine(&result, "BEGIN:VEVENT")
		writeProperty(&result, "UID", event.UID)
		writeCalendarLine(&result, "DTSTAMP:"+dtstamp)
		writeCalendarLine(&result, "DTSTART:"+moscowWallTimeUTC(event.StartTime))
		writeCalendarLine(&result, "DTEND:"+moscowWallTimeUTC(event.EndTime))
		writeProperty(&result, "SUMMARY", summary)
		writeCalendarLine(&result, "CATEGORIES:EDUCATION,"+lessonCategory(event.LessonType))
		writeProperty(&result, "DESCRIPTION", eventDescription(event, lessonTypeName))
		writeProperty(&result, "LOCATION", eventLocation(auditoriums))
		writeCalendarLine(&result, "GEO:"+calendarGeo)
		writeCalendarLine(&result, "URL:"+calendarURL)
		writeCalendarLine(&result, "SEQUENCE:"+strconv.FormatInt(event.Sequence+1, 10))
		writeCalendarLine(&result, "STATUS:CONFIRMED")
		writeCalendarLine(&result, "TRANSP:OPAQUE")
		writeCalendarLine(&result, "BEGIN:VALARM")
		writeCalendarLine(&result, "TRIGGER:-PT20M")
		writeCalendarLine(&result, "ACTION:DISPLAY")
		writeProperty(&result, "DESCRIPTION", alarmDescription)
		writeCalendarLine(&result, "END:VALARM")
		writeCalendarLine(&result, "END:VEVENT")
	}

	writeCalendarLine(&result, "END:VCALENDAR")
	return []byte(result.String())
}

func lessonCategory(lessonType string) string {
	category, _, _ := strings.Cut(lessonType, " ")
	return strings.ToUpper(category)
}

func moscowWallTimeUTC(value time.Time) string {
	const moscowOffset = 3 * time.Hour
	return time.Date(
		value.Year(), value.Month(), value.Day(),
		value.Hour(), value.Minute(), value.Second(), 0,
		time.UTC,
	).Add(-moscowOffset).Format("20060102T150405Z")
}

func eventDescription(event *models.CalendarEvent, lessonTypeName string) string {
	lines := []string{lessonTypeName, event.Title}
	pairs := append([]models.CalendarTeacherAuditorium(nil), event.TeacherAuditoriums...)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Teacher+"\x00"+pairs[i].Auditorium < pairs[j].Teacher+"\x00"+pairs[j].Auditorium
	})
	for _, pair := range pairs {
		teacher, auditorium := strings.TrimSpace(pair.Teacher), strings.TrimSpace(pair.Auditorium)
		switch {
		case teacher != "" && auditorium != "":
			lines = append(lines, teacher+" — "+auditorium)
		case teacher != "":
			lines = append(lines, teacher)
		case auditorium != "":
			lines = append(lines, auditorium)
		}
	}
	if breakPeriod := getBreakPeriod(event.StartTime, event.EndTime); breakPeriod != "" {
		lines = append(lines, "", "Перерыв: "+breakPeriod)
	}
	return strings.Join(lines, "\n")
}

func eventAuditoriums(event *models.CalendarEvent) []string {
	auditoriums := make([]string, 0, len(event.TeacherAuditoriums))
	for _, pair := range event.TeacherAuditoriums {
		auditoriums = append(auditoriums, pair.Auditorium)
	}
	return uniqueSorted(auditoriums)
}

func eventLocation(auditoriums []string) string {
	auditoriums = uniqueSorted(auditoriums)
	if len(auditoriums) == 0 {
		return "РГРТУ"
	}
	return strings.Join(auditoriums, ", ") + " · РГРТУ"
}

func uniqueSorted(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			unique[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func lessonPresentation(lessonType string) (emoji, name, shortName string) {
	presentations := [...]struct {
		lessonType string
		emoji      string
		name       string
		shortName  string
	}{
		{"lecture", "📘", "Лекция", "Лек."},
		{"practice", "✏️", "Практика", "Упр."},
		{"lab", "🧪", "Лабораторная работа", "Лаб."},
		{"coursework", "📄", "Курсовая работа", "Курс. раб."},
		{"course_project", "🛠️", "Курсовой проект", "Курс. пр."},
		{"exam", "🎓", "Экзамен", "Экз."},
		{"zachet", "🎓", "Зачёт", "Зач."},
		{"consultation", "❓", "Консультация", "Конс."},
		{"elective", "🧭", "Факультатив", "Фак."},
	}
	for index := range presentations {
		presentation := &presentations[index]
		if presentation.lessonType == lessonType {
			return presentation.emoji, presentation.name, presentation.shortName
		}
	}
	return "🎓", "Занятие", "Зан."
}

func getBreakPeriod(start, end time.Time) string {
	return map[string]string{
		"08:10-09:45": "08:55–09:00", "09:55-11:30": "10:40–10:45",
		"11:40-13:15": "12:25–12:30", "13:35-15:10": "14:20–14:25",
		"15:20-16:55": "16:05–16:10", "17:05-18:40": "17:50–17:55",
		"18:50-20:15": "19:35–19:40", "20:25-21:50": "21:10–21:15",
	}[start.Format("15:04")+"-"+end.Format("15:04")]
}

func writeProperty(result *strings.Builder, name, value string) {
	writeCalendarLine(result, fmt.Sprintf("%s:%s", name, escapeCalendarText(value)))
}

func escapeCalendarText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, ";", "\\;")
	value = strings.ReplaceAll(value, ",", "\\,")
	value = strings.ReplaceAll(value, "\r\n", "\\n")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.ReplaceAll(value, "\r", "\\n")
}

func writeCalendarLine(result *strings.Builder, line string) {
	const maxOctets = 75
	lineOctets := 0
	for _, character := range line {
		characterOctets := utf8.RuneLen(character)
		if lineOctets+characterOctets > maxOctets {
			result.WriteString("\r\n ")
			lineOctets = 1
		}
		result.WriteRune(character)
		lineOctets += characterOctets
	}
	result.WriteString("\r\n")
}
