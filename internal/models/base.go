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
	Type        string `json:"type"        example:"lab"`
	Description string `json:"description" example:"лабораторная"`
}
