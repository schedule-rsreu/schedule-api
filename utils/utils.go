package utils

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"time"
)

type WeekTypeCash struct {
	WeekType string
	date     time.Time
}

var weekTypeCash = &WeekTypeCash{}

func GetWeekType() (string, error) {
	if weekTypeCash.WeekType != "" && weekTypeCash.date.Day() == time.Now().Day() {
		return weekTypeCash.WeekType, nil
	}
	res, err := http.Get("https://edu.rsreu.ru/")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	weekType := doc.Find(".item.date div:nth-child(2) span").Text()

	weekTypeCash.WeekType = weekType
	weekTypeCash.date = time.Now()

	return weekType, nil
}
