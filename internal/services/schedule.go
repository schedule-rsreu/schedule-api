package services

import (
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"github.com/schedule-rsreu/schedule-api/internal/repo"
	utils "github.com/schedule-rsreu/schedule-api/internal/utils"
)

type ScheduleService struct {
	Repo *repo.ScheduleRepo
}

func NewScheduleService(repo *repo.ScheduleRepo) *ScheduleService {
	return &ScheduleService{
		Repo: repo,
	}
}

func (s *ScheduleService) GetScheduleByGroup(group string) (*models.Schedule, error) {
	group = strings.ToLower(group)

	resp, err := s.Repo.GetScheduleByGroup(group)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedule for group %v not found", group)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetSchedulesByGroups(groups []string) ([]*models.Schedule, error) {
	resp, err := s.Repo.GetSchedulesByGroups(groups)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedules for groups `%v` not found", groups)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetGroups(facultyName string, course int) (*models.CourseFacultyGroups, error) {
	resp, err := s.Repo.GetGroups(facultyName, course)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("groups for faculty %v and course %v not found", facultyName, course)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetFaculties() (*models.Faculties, error) {
	resp, err := s.Repo.GetFaculties()
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{"faculties not found"}
		}
		return nil, err
	}
	return resp, err
}
func (s *ScheduleService) GetFacultyCourses(facultyName string) (*models.FacultyCourses, error) {
	resp, err := s.Repo.GetFacultyCourses(facultyName)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("courses for faculty %v not found", facultyName)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetCourseFaculties(course int) (*models.CourseFaculties, error) {
	resp, err := s.Repo.GetCourseFaculties(course)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("faculties for course %v not found", course)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetDay() (*models.Day, error) {
	w := utils.GetWeekType()

	loc, err := time.LoadLocation("Europe/Moscow")

	if err != nil {
		return nil, err
	}
	now := time.Now().In(loc)

	shortDayNames := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

	WeekTypeMap := map[string]string{
		"числитель":   "numerator",
		"знаменатель": "denominator",
	}

	return &models.Day{
		WeekType:    w,
		WeekTypeEng: WeekTypeMap[w],
		Day:         now.Weekday().String(),
		DayRu:       shortDayNames[now.Weekday()],
		Time:        now.Format("15.04"),
	}, nil
}

func (s *ScheduleService) GetTeacherSchedule(teacher string) (*models.TeacherSchedule, error) {
	resp, err := s.Repo.GetTeacherSchedule(teacher)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedule for teacher '%v' not found", teacher)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetTeachers() (*models.TeachersList, error) {
	return s.Repo.GetAllTeachers()
}
