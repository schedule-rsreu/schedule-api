package utils

import (
	"time"
	_ "time/tzdata"
)

type WeekTypeCash struct {
	WeekType string
	date     time.Time
}

var weekTypeCash = &WeekTypeCash{}
var loc, _ = time.LoadLocation("Europe/Moscow")
var startDate = time.Date(2024, 9, 2, 0, 0, 0, 0, loc)

func GetWeekType() string {

	now := time.Now().In(loc)

	if weekTypeCash.WeekType != "" && weekTypeCash.date.Day() == now.Day() {
		return weekTypeCash.WeekType
	}

	r := now.Sub(startDate).Hours() / 24

	var res string
	if int(r/7)%2 == 0 {
		res = "числитель"
	} else {
		res = "знаменатель"
	}
	weekTypeCash.WeekType = res
	weekTypeCash.date = now
	return res
}
