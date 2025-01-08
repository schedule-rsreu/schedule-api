package models

type TeacherLesson struct {
	Time      string   `json:"time"      bson:"time"      example:"08.10-09.45"`
	Lesson    string   `json:"lesson"    bson:"lesson"    example:"л.Высшая математика\nдоц.Конюхов А.Н.   333 С"`
	Faculties []string `json:"faculties" bson:"faculties" example:"фаиту,фвт"`
	Groups    []string `json:"groups"    bson:"groups"    example:"344,345"`
	Courses   []int    `json:"courses"   bson:"courses"   example:"1"`
}

type TeacherSchedule struct {
	TeacherFull     string `json:"teacher"          bson:"teacher_full"     example:"Конюхов Алексей Николаевич"`
	TeacherShort    string `json:"teacher_short"    bson:"teacher_short"    example:"Конюхов А.Н."`
	Link            string `json:"link"             bson:"link"             example:"https://rsreu.ru/faculties/faitu/kafedri/vm/prepodavateli/9402-item-9402"` //nolint:lll // there is no way to fix it
	Department      string `json:"department"       bson:"department"       example:"Кафедра высшей математики"`
	DepartmentShort string `json:"department_short" bson:"department_short" example:"ВМ"`
	Faculty         string `json:"faculty"          bson:"faculty"          example:"Факультет автоматики и информационных технологий в управлении"` //nolint:lll // there is no way to fix it
	FacultyShort    string `json:"faculty_short"    bson:"faculty_short"    example:"фаиту"`
	Schedule        struct {
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

type TeacherFaculty struct {
	Faculty      string `json:"faculty"       bson:"faculty"       example:"Факультет автоматики и информационных технологий в управлении"` //nolint:lll // there is no way to fix it
	FacultyShort string `json:"faculty_short" bson:"faculty_short" example:"фаиту"`
}

type TeacherDepartment struct {
	Department      string `json:"department"       bson:"department"       example:"Кафедра вычислительной и прикладной математики"` //nolint:lll // there is no way to fix it
	DepartmentShort string `json:"department_short" bson:"department_short" example:"ВМ"`
}
