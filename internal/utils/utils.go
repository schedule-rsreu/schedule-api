package utils

import (
	"time"
	_ "time/tzdata"
)

const timeZone = "Europe/Moscow"

func GetWeekType() string {
	const dayHours = 24
	const weekDays = 7

	var loc, err = time.LoadLocation(timeZone)

	if err != nil {
		return ""
	}

	var startDate = time.Date(2025, 8, 18, 0, 0, 0, 0, loc)

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

func GetWeekBounds(date time.Time) (time.Time, time.Time) {
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday as 7
	}
	// понедельник этой недели
	monday := date.AddDate(0, 0, -weekday+1)
	// воскресенье следующей недели (через 13 дней от monday)
	nextSunday := monday.AddDate(0, 0, 13)
	return monday, nextSunday
}
