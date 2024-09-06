package services

import (
	"github.com/VinGP/schedule-api/repo"
	"github.com/VinGP/schedule-api/scheme"
	utils "github.com/VinGP/schedule-api/utils"
	"time"
	_ "time/tzdata"
)

type ScheduleService struct {
	Repo *repo.ScheduleRepo
}

func (s *ScheduleService) GetScheduleByGroup(group string) (*scheme.Schedule, error) {
	return s.Repo.GetScheduleByGroup(group)
}

func (s *ScheduleService) GetSchedulesByGroups(groups []string) ([]*scheme.Schedule, error) {
	return s.Repo.GetSchedulesByGroups(groups)
}

func (s *ScheduleService) GetGroups(facultyName string, course int) (scheme.CourseFacultyGroups, error) {
	return s.Repo.GetGroups(facultyName, course)
}

func (s *ScheduleService) GetFaculties() (scheme.Faculties, error) {
	return s.Repo.GetFaculties()
}
func (s *ScheduleService) GetFacultyCourses(facultyName string) (scheme.FacultyCourses, error) {
	return s.Repo.GetFacultyCourses(facultyName)
}

func (s *ScheduleService) GetDay() (scheme.Day, error) {
	w, err := utils.GetWeekType()

	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc)

	if err != nil {
		return scheme.Day{}, err
	}

	shortDayNames := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

	WeekTypeMap := map[string]string{
		"числитель":   "numerator",
		"знаменатель": "denominator",
	}

	return scheme.Day{
		WeekType:    w,
		WeekTypeEng: WeekTypeMap[w],
		Day:         now.Weekday().String(),
		DayRu:       shortDayNames[now.Weekday()],
		Time:        now.Format("15:04"),
	}, nil
}
