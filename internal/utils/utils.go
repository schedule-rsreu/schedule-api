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

	var startDate = time.Date(2024, 9, 2, 0, 0, 0, 0, loc)

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
