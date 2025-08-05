package models

type StudentLesson struct {
	Time               string                     `json:"time"                bson:"time"                example:"08.10-09.45"`
	Lesson             string                     `json:"lesson"              bson:"lesson"              example:"Высшая математика"` //nolint:lll // there is no way to fix it
	Date               string                     `json:"date"                bson:"date"                example:"2025-06-18"`
	Type               string                     `json:"type"                bson:"type"                example:"lab,practice"`
	StartTime          string                     `json:"start_time"          bson:"start_time"          example:"2025-06-18T15:20:00"`
	EndTime            string                     `json:"end_time"            bson:"end_time"            example:"2025-06-18T16:55:00"`
	TeacherAuditoriums []StudentTeacherAuditorium `json:"teacher_auditoriums" bson:"teacher_auditoriums"`
}

type StudentWeek Week[StudentLesson]

type StudentTeacherAuditorium struct {
	Teacher    *StudentTeacherInfo `json:"teacher"`
	Auditorium *Auditorium         `json:"auditorium"`
}

type StudentTeacherInfo struct {
	FullName  string `json:"full_name"  example:"Конюхов Алексей Николаевич"`
	ShortName string `json:"short_name" example:"Конюхов А.Н."`
	Id        int    `json:"id"         example:"1"`
}

type StudentSchedule struct {
	Faculty           string                            `json:"faculty"            bson:"faculty"            example:"фвт"`
	Group             string                            `json:"group"              bson:"group"              example:"344"`
	Course            int                               `json:"course"             bson:"course"             example:"1"`
	NumeratorPeriod   string                            `json:"numerator_period"   bson:"numerator_period"   example:"16.06-22.06"`
	DenominatorPeriod string                            `json:"denominator_period" bson:"denominator_period" example:"09.06-15.06"`
	InputWeekType     string                            `json:"input_week_type"`
	Schedule          NumeratorDenominator[StudentWeek] `json:"schedule"           bson:"schedule"`
	LessonsTimes      []string                          `json:"lessons_times,omitempty"                  db:"lessons_times"`
}
