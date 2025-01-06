package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DayLessonSchedule struct {
	Time          string   `json:"time"           bson:"time"           example:"08.10-09.45"`
	Lesson        string   `json:"lesson"         bson:"lesson"         example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"` //nolint:lll // there is no way to fix it
	TeachersFull  []string `json:"teachers_full"  bson:"teachers_full"  example:"Конюхов Алексей Николаевич"`
	TeachersShort []string `json:"teachers_short" bson:"teachers_short" example:"Конюхов А.Н."`
}

type WeekSchedule struct {
	Monday    []DayLessonSchedule `json:"monday"    bson:"monday"`
	Tuesday   []DayLessonSchedule `json:"tuesday"   bson:"tuesday"`
	Wednesday []DayLessonSchedule `json:"wednesday" bson:"wednesday"`
	Thursday  []DayLessonSchedule `json:"thursday"  bson:"thursday"`
	Friday    []DayLessonSchedule `json:"friday"    bson:"friday"`
	Saturday  []DayLessonSchedule `json:"saturday"  bson:"saturday"`
}

type NumeratorDenominatorSchedule struct {
	Numerator   WeekSchedule `json:"numerator"   bson:"numerator"`
	Denominator WeekSchedule `json:"denominator" bson:"denominator"`
}

type Schedule struct {
	UpdatedAt time.Time                    `json:"updated_at" bson:"updated_at"`
	File      string                       `json:"file"       bson:"file"`
	FileHash  string                       `json:"file_hash"  bson:"file_hash"  example:"5427593514859b0701e8e12ecbce1b0b"`
	Faculty   string                       `json:"faculty"    bson:"faculty"    example:"фвт"`
	Group     string                       `json:"group"      bson:"group"      example:"344"`
	Schedule  NumeratorDenominatorSchedule `json:"schedule"   bson:"schedule"`
	Course    int                          `json:"course"     bson:"course"     example:"1"`
	ID        primitive.ObjectID           `json:"id"         bson:"_id"`
}

type CourseFacultyGroups struct {
	Faculty string   `json:"faculty" bson:"faculty" enums:"иэф,фаиту,фвт,фрт,фэ"                 example:"фвт"`
	Groups  []string `json:"groups"  bson:"groups"`
	Course  int      `json:"course"  bson:"course"  enums:"1,2,3,4,5"                            example:"1"`
}

type Faculties struct {
	Faculties []string `json:"faculties" bson:"faculties"`
}

type FacultyCourses struct {
	Faculty string `json:"faculty" bson:"faculty" enums:"иэф,фаиту,фвт,фрт,фэ"                 example:"фвт"`
	Courses []int  `json:"courses" bson:"courses"`
}

type CourseFaculties struct {
	Faculties []string `json:"faculties" bson:"faculties" enums:"иэф,фаиту,фвт,фрт,фэ"`
	Course    int      `json:"course"    bson:"course"    enums:"1,2,3,4,5"                            example:"1"`
}

type Day struct {
	WeekType    string `json:"week_type"     enums:"числитель,знаменатель"                                    example:"знаменатель"` //nolint:lll // there is no way to fix it
	WeekTypeEng string `json:"week_type_eng" enums:"numerator,denominator"                                    example:"numerator"`   //nolint:lll // there is no way to fix it
	Day         string `json:"day"           enums:"Monday,Tuesday,Wednesday,Thursday,Friday,Saturday,Sunday" example:"Monday"`      //nolint:lll // there is no way to fix it
	DayRu       string `json:"day_ru"        enums:"Пн,Вт,Ср,Чт,Пт,Сб,Вс"                                     example:"Пн"`          //nolint:lll // there is no way to fix it
	Time        string `json:"time"                                                                           example:"08.10"`       //nolint:lll // there is no way to fix it
}
