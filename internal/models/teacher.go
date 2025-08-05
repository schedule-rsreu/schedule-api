package models

type TeacherLesson struct {
	Time       string     `json:"time"        bson:"time"        example:"08.10-09.45"`
	Lesson     string     `json:"lesson"      bson:"lesson"      example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"`
	Type       string     `json:"type"        bson:"type"        example:"lab,practice"`
	Date       string     `json:"date"        bson:"date"        example:"2025-06-18"`
	Faculties  []string   `json:"faculties"   bson:"faculties"   example:"фаиту,фвт"`
	Groups     []string   `json:"groups"      bson:"groups"      example:"344,345"`
	Courses    []int      `json:"courses"     bson:"courses"     example:"1"`
	Auditorium Auditorium `json:"auditorium" bson:"auditorium" `
}

type TeacherWeek Week[TeacherLesson]

type TeacherSchedule struct {
	Id          int          `json:"id" bson:"id" example:"1"`
	FullName    string       `json:"full_name"              example:"Конюхов Алексей Николаевич"`
	ShortName   string       `json:"short_name"      example:"Конюхов А.Н."`
	Link        string       `json:"link"             bson:"link"             example:"https://rsreu.ru/faculties/faitu/kafedri/vm/prepodavateli/9402-item-9402"` //nolint:lll // there is no way to fix it
	Departments []Department `json:"departments" bson:"departments"`

	NumeratorPeriod   string `json:"numerator_period"   bson:"numerator_period"   example:"16.06-22.06"`
	DenominatorPeriod string `json:"denominator_period" bson:"denominator_period" example:"09.06-15.06"`
	InputWeekType     string `json:"input_week_type"`

	Schedule     NumeratorDenominator[TeacherWeek] `json:"schedule" bson:"schedule"`
	LessonsTimes []string                          `json:"lessons_times,omitempty"                  db:"lessons_times"`
}

type TeacherInfo struct {
	Id          int          `json:"id" bson:"id" example:"1"`
	FullName    string       `json:"full_name"              example:"Конюхов Алексей Николаевич"`
	ShortName   string       `json:"short_name"      example:"Конюхов А.Н."`
	Link        string       `json:"link"             bson:"link"             example:"https://rsreu.ru/faculties/faitu/kafedri/vm/prepodavateli/9402-item-9402"` //nolint:lll // there is no way to fix it
	Departments []Department `json:"departments" bson:"departments"`
}

type TeachersList struct {
	Teachers []StudentTeacherInfo `json:"teachers" bson:"teachers" `
}
