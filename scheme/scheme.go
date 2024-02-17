package scheme

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DayLessonSchedule struct {
	Time   string `bson:"time" json:"time"`
	Lesson string `bson:"lesson" json:"lesson"`
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
	Faculty  string                       `bson:"faculty" json:"faculty"`
	Course   int                          `bson:"course" json:"course"`
	Group    string                       `bson:"group" json:"group"`
	Schedule NumeratorDenominatorSchedule `bson:"schedule" json:"schedule"`
}

type CourseFacultyGroups struct {
	Faculty string   `bson:"faculty" json:"faculty"`
	Course  int      `bson:"course" json:"course"`
	Groups  []string `bson:"groups" json:"groups"`
}

type Faculties struct {
	Faculties []string `bson:"faculties" json:"faculties"`
}

type FacultyCourses struct {
	Faculty string `bson:"faculty" json:"faculty"`
	Courses []int  `bson:"courses" json:"courses"`
}
