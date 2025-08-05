package models

type Auditorium struct {
	Building    Building `json:"building"`
	Number      string   `json:"number"       example:"445"`
	DisplayName string   `json:"display_name" example:"445 C"`
	Id          int      `json:"id"           example:"1"`
}

type AuditoriumLesson struct {
	Time      string               `json:"time"        bson:"time"        example:"08.10-09.45"`
	Lesson    string               `json:"lesson"      bson:"lesson"      example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"`
	Faculties []string             `json:"faculties"   bson:"faculties"   example:"фаиту,фвт"`
	Groups    []string             `json:"groups"      bson:"groups"      example:"344,345"`
	Courses   []int                `json:"courses"     bson:"courses"     example:"1"`
	Teachers  []StudentTeacherInfo `json:"teachers" bson:"teachers" `
}

type AuditoriumWeek Week[AuditoriumLesson]

type AuditoriumSchedule struct {
	Auditorium Auditorium                           `json:"auditorium" bson:"auditorium"`
	Schedule   NumeratorDenominator[AuditoriumWeek] `json:"schedule" bson:"schedule"`
}

type AuditoriumsList []Auditorium
