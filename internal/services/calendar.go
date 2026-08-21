package services

import (
	"context"
	"errors"
	"fmt"
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

func (s *ScheduleService) GetGroupCalendar(ctx context.Context, group, source string) ([]byte, error) {
	group = strings.ToUpper(strings.TrimSpace(group))
	calendar, err := s.Repo.GetGroupCalendar(ctx, group)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("calendar for group %v not found", group)}
		}
		return nil, err
	}
	calendar.Source = source
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
	if calendar.Source != "" {
		writeCalendarLine(&result, "SOURCE;VALUE=URI:"+calendar.Source)
	}
	writeCalendarLine(&result, "REFRESH-INTERVAL;VALUE=DURATION:PT1H")
	writeCalendarLine(&result, "COLOR:royalblue")

	dtstamp := calendar.UpdatedAt.UTC().Format("20060102T150405Z")
	for index := range calendar.Events {
		event := &calendar.Events[index]
		emoji, lessonTypeName := lessonPresentation(event.LessonType)
		writeCalendarLine(&result, "BEGIN:VEVENT")
		writeProperty(&result, "UID", event.UID)
		writeCalendarLine(&result, "DTSTAMP:"+dtstamp)
		writeCalendarLine(&result, "DTSTART:"+moscowWallTimeUTC(event.StartTime))
		writeCalendarLine(&result, "DTEND:"+moscowWallTimeUTC(event.EndTime))
		writeProperty(&result, "SUMMARY", emoji+" "+event.Title)
		writeCalendarLine(&result, "CATEGORIES:EDUCATION,"+lessonCategory(event.LessonType))
		writeProperty(&result, "DESCRIPTION", eventDescription(event, emoji, lessonTypeName))
		writeProperty(&result, "LOCATION", eventLocation(event.Auditoriums))
		writeCalendarLine(&result, "GEO:"+calendarGeo)
		writeCalendarLine(&result, "URL:"+calendarURL)
		writeCalendarLine(&result, "SEQUENCE:"+strconv.FormatInt(event.Sequence, 10))
		if event.Cancelled {
			writeCalendarLine(&result, "STATUS:CANCELLED")
		} else {
			writeCalendarLine(&result, "STATUS:CONFIRMED")
		}
		writeCalendarLine(&result, "TRANSP:OPAQUE")
		if !event.Cancelled {
			writeCalendarLine(&result, "BEGIN:VALARM")
			writeCalendarLine(&result, "TRIGGER:-PT30M")
			writeCalendarLine(&result, "ACTION:DISPLAY")
			writeProperty(&result, "DESCRIPTION", "Пара через 30 минут: "+event.Title)
			writeCalendarLine(&result, "END:VALARM")
		}
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

func eventDescription(event *models.CalendarEvent, emoji, lessonTypeName string) string {
	description := emoji + " " + lessonTypeName + " " + event.Title
	teachers := uniqueSorted(event.Teachers)
	if len(teachers) > 0 {
		return description + "\n🧑‍🏫 " + strings.Join(teachers, ", ")
	}
	return description
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

func lessonPresentation(lessonType string) (emoji, name string) {
	presentations := [...]struct {
		lessonType string
		emoji      string
		name       string
	}{
		{"lecture", "📘", "Лекция"},
		{"practice", "✏️", "Практика"},
		{"lab", "🧪", "Лабораторная работа"},
		{"coursework", "📄", "Курсовая работа"},
		{"course_project", "🛠️", "Курсовой проект"},
		{"exam", "🎓", "Экзамен"},
		{"zachet", "🎓", "Зачёт"},
		{"consultation", "❓", "Консультация"},
		{"elective", "🧭", "Факультатив"},
	}
	for _, presentation := range presentations {
		if presentation.lessonType == lessonType {
			return presentation.emoji, presentation.name
		}
	}
	return "🎓", "Занятие"
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
