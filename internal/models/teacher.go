package models

type TeacherLesson struct {
	Time      string   `json:"time"      bson:"time"      example:"08.10-09.45"`
	Lesson    string   `json:"lesson"    bson:"lesson"    example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"`
	Faculties []string `json:"faculties" bson:"faculties" example:"фаиту,фвт"`
	Groups    []string `json:"groups"    bson:"groups"    example:"344,345"`
	Courses   []int    `json:"courses"   bson:"courses"   example:"1"`
}

type TeacherSchedule struct {
	TeacherFull string `json:"teacher" bson:"teacher_full" example:"Конюхов Алексей Николаевич,Маношкин Алексей Борисович"`
	Schedule    struct {
		Numerator struct {
			Monday    []TeacherLesson `json:"monday"    bson:"monday"`
			Tuesday   []TeacherLesson `json:"tuesday"   bson:"tuesday"`
			Wednesday []TeacherLesson `json:"wednesday" bson:"wednesday"`
			Thursday  []TeacherLesson `json:"thursday"  bson:"thursday"`
			Friday    []TeacherLesson `json:"friday"    bson:"friday"`
			Saturday  []TeacherLesson `json:"saturday"  bson:"saturday"`
		} `json:"numerator" bson:"numerator"`
		Denominator struct {
			Monday    []TeacherLesson `json:"monday"    bson:"monday"`
			Tuesday   []TeacherLesson `json:"tuesday"   bson:"tuesday"`
			Wednesday []TeacherLesson `json:"wednesday" bson:"wednesday"`
			Thursday  []TeacherLesson `json:"thursday"  bson:"thursday"`
			Friday    []TeacherLesson `json:"friday"    bson:"friday"`
			Saturday  []TeacherLesson `json:"saturday"  bson:"saturday"`
		} `json:"denominator" bson:"denominator"`
	} `json:"schedule" bson:"schedule"`
}

type TeachersList struct {
	Teachers []string `json:"teachers" bson:"teachers" example:"Конюхов Алексей Николаевич,Маношкин Алексей Борисович"`
}
