package models

type Week[TLesson StudentLesson | TeacherLesson | AuditoriumLesson] struct {
	Monday    []TLesson `json:"monday"    bson:"monday"`
	Tuesday   []TLesson `json:"tuesday"   bson:"tuesday"`
	Wednesday []TLesson `json:"wednesday" bson:"wednesday"`
	Thursday  []TLesson `json:"thursday"  bson:"thursday"`
	Friday    []TLesson `json:"friday"    bson:"friday"`
	Saturday  []TLesson `json:"saturday"  bson:"saturday"`
}

type NumeratorDenominator[TWeek StudentWeek | TeacherWeek | AuditoriumWeek] struct {
	Numerator   TWeek `json:"numerator"   bson:"numerator"`
	Denominator TWeek `json:"denominator" bson:"denominator"`
}

type LessonType struct {
	Type        string `json:"type" example:"lab"`
	Description string `json:"description" example:"лабораторная"`
}

var LessonTypes = []LessonType{
	{Type: "lecture", Description: "лекция"},
	{Type: "lab", Description: "лабораторная"},
	{Type: "practice", Description: "практика"},
	{Type: "coursework", Description: "курсовая работа"},
	{Type: "course_project", Description: "курсовой проект"},
	{Type: "exam", Description: "экзамен"},
	{Type: "zachet", Description: "зачет"},
	{Type: "consultation", Description: "консультация"},
	{Type: "elective", Description: "факультатив"},
	{Type: "unknown", Description: ""},
}
