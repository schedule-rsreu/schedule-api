package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"github.com/schedule-rsreu/schedule-api/internal/repo"
	"github.com/schedule-rsreu/schedule-api/internal/utils"
)

type ScheduleService struct {
	Repo *repo.ScheduleRepo
}

func NewScheduleService(scheduleRepo *repo.ScheduleRepo) *ScheduleService {
	return &ScheduleService{
		Repo: scheduleRepo,
	}
}

func (s *ScheduleService) GetScheduleByGroup(group string, addEmptyLessons bool) (*models.Schedule, error) {
	group = strings.ToLower(group)

	resp, err := s.Repo.GetScheduleByGroup(group)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedule for group %v not found", group)}
		}
		return nil, err
	}

	if addEmptyLessons {
		s.AddEmptyLessons(&resp.Schedule)
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

func (s *ScheduleService) GetAllTeachers() (*models.TeachersList, error) {
	return s.Repo.GetAllTeachers()
}

func (s *ScheduleService) GetTeachersList(faculty, department *string) (*models.TeachersList, error) {
	resp, err := s.Repo.GetTeachersList(faculty, department)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{
				fmt.Sprintf(
					"teachers for faculty '%v' and department '%v' not found",
					*faculty, *department)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetTeachersFaculties(department *string) ([]*models.TeacherFaculty, error) {
	resp, err := s.Repo.GetTeachersFaculties(department)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("faculties for department '%v' not found", *department)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetTeachersDepartments(faculty *string) ([]*models.TeacherDepartment, error) {
	resp, err := s.Repo.GetTeachersDepartments(faculty)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("departments for faculty '%v' not found", *faculty)}
		}
		return nil, err
	}
	return resp, err
}

func addEmptyLessons(lessons []models.DayLessonSchedule, times []string) []models.DayLessonSchedule {
	if len(times) == 0 {
		return lessons
	}

	const emptyLessonText = "—"

	existingTimes := make(map[string]struct{})
	for i := range lessons {
		lesson := &lessons[i]
		existingTimes[lesson.Time] = struct{}{}
	}

	for _, time := range times {
		if _, exists := existingTimes[time]; !exists {
			lessons = append(lessons, models.DayLessonSchedule{
				Time:          time,
				Lesson:        emptyLessonText,
				TeachersFull:  []string{},
				TeachersShort: []string{},
			})
		}
	}

	sort.Slice(lessons, func(i, j int) bool {
		return lessons[i].Time < lessons[j].Time
	})

	for i := len(lessons) - 1; i >= 0; i-- {
		if lessons[i].Lesson == emptyLessonText {
			lessons = lessons[:i]
		} else {
			break
		}
	}

	return lessons
}

func (s *ScheduleService) GetFacultiesWithCourses() (*models.FacultiesCourses, error) {
	resp, err := s.Repo.GetFacultiesWithCourses()

	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{"no results"}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) AddEmptyLessons(schedule *models.NumeratorDenominatorSchedule) {
	times := schedule.Denominator.WeekDayLessonsTimes

	schedule.Denominator.Monday = addEmptyLessons(schedule.Denominator.Monday, times)
	schedule.Denominator.Tuesday = addEmptyLessons(schedule.Denominator.Tuesday, times)
	schedule.Denominator.Wednesday = addEmptyLessons(schedule.Denominator.Wednesday, times)
	schedule.Denominator.Thursday = addEmptyLessons(schedule.Denominator.Thursday, times)
	schedule.Denominator.Friday = addEmptyLessons(schedule.Denominator.Friday, times)
	schedule.Denominator.Saturday = addEmptyLessons(schedule.Denominator.Saturday, times)

	times = schedule.Numerator.WeekDayLessonsTimes

	schedule.Numerator.Monday = addEmptyLessons(schedule.Numerator.Monday, times)
	schedule.Numerator.Tuesday = addEmptyLessons(schedule.Numerator.Tuesday, times)
	schedule.Numerator.Wednesday = addEmptyLessons(schedule.Numerator.Wednesday, times)
	schedule.Numerator.Thursday = addEmptyLessons(schedule.Numerator.Thursday, times)
	schedule.Numerator.Friday = addEmptyLessons(schedule.Numerator.Friday, times)
	schedule.Numerator.Saturday = addEmptyLessons(schedule.Numerator.Saturday, times)
}
