package utils

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"time"
	_ "time/tzdata"
)

type WeekTypeCash struct {
	WeekType string
	date     time.Time
}

var weekTypeCash = &WeekTypeCash{}

func GetWeekType() (string, error) {
	return "числитель", nil
	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc)

	cashWeekType := weekTypeCash.WeekType

	if weekTypeCash.WeekType != "" && weekTypeCash.date.Day() == now.Day() {
		return cashWeekType, nil
	}
	res, err := http.Get("https://edu.rsreu.ru/")
	if err != nil {
		return cashWeekType, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return cashWeekType, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return cashWeekType, err
	}
	weekType := doc.Find(".item.date div:nth-child(2) span").Text()

	weekTypeCash.WeekType = weekType
	weekTypeCash.date = now

	return weekType, nil
}
