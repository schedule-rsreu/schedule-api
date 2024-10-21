package utils

import (
	"time"
	_ "time/tzdata"
)

type WeekTypeCash struct {
	WeekType string
	date     time.Time
}

var loc, _ = time.LoadLocation("Europe/Moscow")
var startDate = time.Date(2024, 9, 2, 0, 0, 0, 0, loc)

func GetWeekType() string {

	date := time.Now().In(loc)

	r := date.Sub(startDate).Hours() / 24

	if int(r/7)%2 == 0 {
		return "числитель"
	} else {
		return "знаменатель"
	}
}
