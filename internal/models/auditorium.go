package models

type Auditorium struct {
	Building    Building `json:"building"`
	Number      string   `json:"number"       example:"445"`
	DisplayName string   `json:"display_name" example:"445 C"`
	Id          int      `json:"id"           example:"1"`
}

type AuditoriumLesson struct {
	Time      string               `json:"time"        bson:"time"        example:"08.10-09.45"`
	Date      string               `json:"date"        bson:"date"        example:"2025-06-18"`
	Type      string               `json:"type"        bson:"type"        example:"lab,practice"`
	Lesson    string               `json:"lesson"      bson:"lesson"      example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"`
	Faculties []string             `json:"faculties"   bson:"faculties"   example:"фаиту,фвт"`
	Groups    []string             `json:"groups"      bson:"groups"      example:"344,345"`
	Courses   []int                `json:"courses"     bson:"courses"     example:"1"`
	Teachers  []StudentTeacherInfo `json:"teachers" bson:"teachers" `
}

type AuditoriumWeek Week[AuditoriumLesson]

type AuditoriumSchedule struct {
	Auditorium Auditorium `json:"auditorium" bson:"auditorium"`

	NumeratorPeriod   string `json:"numerator_period"   bson:"numerator_period"   example:"16.06-22.06"`
	DenominatorPeriod string `json:"denominator_period" bson:"denominator_period" example:"09.06-15.06"`
	InputWeekType     string `json:"input_week_type" example:"numerator"`

	Schedule NumeratorDenominator[AuditoriumWeek] `json:"schedule" bson:"schedule"`
}

type AuditoriumsList []Auditorium
