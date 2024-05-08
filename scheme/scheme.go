package scheme

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type DayLessonSchedule struct {
	Time   string `bson:"time" json:"time" example:"08.10-09.45"`
	Lesson string `bson:"lesson" json:"lesson" example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"`
}

type WeekSchedule struct {
	Monday    []DayLessonSchedule `bson:"monday" json:"monday"`
	Tuesday   []DayLessonSchedule `bson:"tuesday" json:"tuesday"`
	Wednesday []DayLessonSchedule `bson:"wednesday" json:"wednesday"`
	Thursday  []DayLessonSchedule `bson:"thursday" json:"thursday"`
	Friday    []DayLessonSchedule `bson:"friday" json:"friday"`
	Saturday  []DayLessonSchedule `bson:"saturday" json:"saturday"`
}

type NumeratorDenominatorSchedule struct {
	Numerator   WeekSchedule `bson:"numerator" json:"numerator"`
	Denominator WeekSchedule `bson:"denominator" json:"denominator"`
}

type Schedule struct {
	ID       primitive.ObjectID           `bson:"_id" json:"id"`
	UpdateAt time.Time                    `bson:"update_at" json:"update_at"`
	File     string                       `bson:"file" json:"file"`
	FileHash string                       `bson:"file_hash" json:"file_hash" example:"5427593514859b0701e8e12ecbce1b0b"`
	Faculty  string                       `bson:"faculty" json:"faculty" example:"фвт"`
	Course   int                          `bson:"course" json:"course" example:"1"`
	Group    string                       `bson:"group" json:"group" example:"344"`
	Schedule NumeratorDenominatorSchedule `bson:"schedule" json:"schedule"`
}

type CourseFacultyGroups struct {
	Faculty string   `bson:"faculty" json:"faculty" example:"фвт" enums:"иэф,фаиту,фвт,фрт,фэ"`
	Course  int      `bson:"course" json:"course" example:"1" enums:"1,2,3,4,5"`
	Groups  []string `bson:"groups" json:"groups"`
}

type Faculties struct {
	Faculties []string `bson:"faculties" json:"faculties"`
}

type FacultyCourses struct {
	Faculty string `bson:"faculty" json:"faculty" example:"фвт" enums:"иэф,фаиту,фвт,фрт,фэ"`
	Courses []int  `bson:"courses" json:"courses"`
}

type WeekType struct {
	WeekType string ` json:"week_type" example:"знаменатель" enums:"числитель,знаменатель"`
}
