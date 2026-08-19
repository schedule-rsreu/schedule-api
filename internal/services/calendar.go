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

func (s *ScheduleService) GetGroupCalendar(ctx context.Context, group string) ([]byte, error) {
	group = strings.ToUpper(strings.TrimSpace(group))
	calendar, err := s.Repo.GetGroupCalendar(ctx, group)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("calendar for group %v not found", group)}
		}
		return nil, err
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
	writeProperty(&result, "X-WR-CALNAME", "Расписание группы "+calendar.Group)
	writeCalendarLine(&result, "X-WR-TIMEZONE:Europe/Moscow")

	dtstamp := calendar.UpdatedAt.UTC().Format("20060102T150405Z")
	for index := range calendar.Events {
		event := &calendar.Events[index]
		writeCalendarLine(&result, "BEGIN:VEVENT")
		writeProperty(&result, "UID", event.UID)
		writeCalendarLine(&result, "DTSTAMP:"+dtstamp)
		writeCalendarLine(&result, "DTSTART:"+moscowWallTimeUTC(event.StartTime))
		writeCalendarLine(&result, "DTEND:"+moscowWallTimeUTC(event.EndTime))
		writeProperty(&result, "SUMMARY", lessonEmoji(event.LessonType)+" "+event.Title)
		writeProperty(&result, "DESCRIPTION", eventDescription(event))
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
		writeCalendarLine(&result, "END:VEVENT")
	}

	writeCalendarLine(&result, "END:VCALENDAR")
	return []byte(result.String())
}

func moscowWallTimeUTC(value time.Time) string {
	const moscowOffset = 3 * time.Hour
	return time.Date(
		value.Year(), value.Month(), value.Day(),
		value.Hour(), value.Minute(), value.Second(), 0,
		time.UTC,
	).Add(-moscowOffset).Format("20060102T150405Z")
}

func eventDescription(event *models.CalendarEvent) string {
	description := lessonTypeName(event.LessonType)
	teachers := uniqueSorted(event.Teachers)
	if len(teachers) == 1 {
		return description + "\nПреподаватель: " + teachers[0]
	}
	if len(teachers) > 1 {
		return description + "\nПреподаватели: " + strings.Join(teachers, ", ")
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

func lessonEmoji(lessonType string) string {
	switch lessonType {
	case "lecture":
		return "📘"
	case "practice":
		return "✏️"
	case "lab":
		return "🧪"
	default:
		return "🎓"
	}
}

func lessonTypeName(lessonType string) string {
	switch lessonType {
	case "lecture":
		return "Лекция"
	case "practice":
		return "Практика"
	case "lab":
		return "Лабораторная работа"
	default:
		return "Занятие"
	}
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
