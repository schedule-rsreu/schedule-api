package models

type FacultyCourses struct {
	Faculty string `json:"faculty" bson:"faculty" enums:"иэф,фаиту,фвт,фрт,фэ"                 example:"фвт"`
	Courses []int  `json:"courses" bson:"courses"`
}

type FacultiesCourses []FacultyCourses
