package models

type Day struct {
	WeekType    string `json:"week_type"     enums:"числитель,знаменатель"                                    example:"знаменатель"` //nolint:lll // there is no way to fix it
	WeekTypeEng string `json:"week_type_eng" enums:"numerator,denominator"                                    example:"numerator"`   //nolint:lll // there is no way to fix it
	Day         string `json:"day"           enums:"Monday,Tuesday,Wednesday,Thursday,Friday,Saturday,Sunday" example:"Monday"`      //nolint:lll // there is no way to fix it
	DayRu       string `json:"day_ru"        enums:"Пн,Вт,Ср,Чт,Пт,Сб,Вс"                                     example:"Пн"`          //nolint:lll // there is no way to fix it
	Time        string `json:"time"                                                                           example:"08.10"`       //nolint:lll // there is no way to fix it
}
