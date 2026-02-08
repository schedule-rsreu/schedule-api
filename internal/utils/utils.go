package utils

import (
	"time"
	_ "time/tzdata"
)

const timeZone = "Europe/Moscow"

func GetNowWithZone() time.Time {
	loc, err := time.LoadLocation(timeZone)

	if err != nil {
		return time.Now()
	}
	now := time.Now().In(loc)
	return now
}

func GetWeekType() string {
	const dayHours = 24
	const weekDays = 7

	var loc, err = time.LoadLocation(timeZone)

	if err != nil {
		return ""
	}

	var startDate = time.Date(2026, 2, 8, 0, 0, 0, 0, loc)

	now := time.Now().In(loc)

	r := now.Sub(startDate).Hours() / dayHours

	var res string

	if int(r/weekDays)%2 == 0 {
		res = "числитель"
	} else {
		res = "знаменатель"
	}

	return res
}

func GetWeekBounds(date time.Time) (monday, nextSunday time.Time) {
	const daysUntilNextSunday = 13

	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday as 7
	}
	// понедельник этой недели
	monday = date.AddDate(0, 0, -weekday+1)
	// воскресенье следующей недели (через 13 дней от monday)
	nextSunday = monday.AddDate(0, 0, daysUntilNextSunday)
	return monday, nextSunday
}

// GetDateRangeBounds возвращает диапазон дат ±monthsOffset от указанной даты.
func GetDateRangeBounds(date time.Time, monthsOffset int) (start, end time.Time) {
	start = date.AddDate(0, -monthsOffset, 0)
	end = date.AddDate(0, monthsOffset, 0)
	return start, end
}
