package models

type Faculty struct {
	Id         int    `json:"id" bson:"id" example:"1"`
	Title      string `json:"title" bson:"title" example:"Факультет автоматики и информационных технологий в управлении"`
	TitleShort string `json:"title_short" bson:"short" example:"фаиту"`
}

type CourseFacultyGroups struct {
	Faculty string   `json:"faculty" bson:"faculty" enums:"иэф,фаиту,фвт,фрт,фэ"                 example:"фвт"`
	Groups  []string `json:"groups"  bson:"groups"`
	Course  int      `json:"course"  bson:"course"  enums:"1,2,3,4,5"                            example:"1"`
}

type Faculties struct {
	Faculties []string `json:"faculties" bson:"faculties"`
}

type CourseFaculties struct {
	Faculties []string `json:"faculties" bson:"faculties" enums:"иэф,фаиту,фвт,фрт,фэ"`
	Course    int      `json:"course"    bson:"course"    enums:"1,2,3,4,5"                            example:"1"`
}
