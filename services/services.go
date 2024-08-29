package services

import (
	"github.com/VinGP/schedule-api/repo"
	"github.com/VinGP/schedule-api/scheme"
	utils "github.com/VinGP/schedule-api/utils"
	"time"
)

type ScheduleService struct {
	Repo *repo.ScheduleRepo
}

func (s *ScheduleService) GetScheduleByGroup(group string) (*scheme.Schedule, error) {
	return s.Repo.GetScheduleByGroup(group)
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

	if err != nil {
		return scheme.Day{}, err
	}

	shortDayNames := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}

	return scheme.Day{
		WeekType: w,
		Day:      time.Now().Weekday().String(),
		DayRu:    shortDayNames[time.Now().Weekday()],
	}, nil
}
