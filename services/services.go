package services

import (
	"github.com/VinGP/schedule-api/repo"
	"github.com/VinGP/schedule-api/scheme"
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
