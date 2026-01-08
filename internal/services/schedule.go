package services

import (
	"context"
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

func ParseDateOrNow(dateStr string) (time.Time, error) {
	var date time.Time
	var err error

	if dateStr == "" {
		date = utils.GetNowWithZone()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return time.Time{}, ErrInvalidDateFormat
		}
	}
	return date, nil
}

func (s *ScheduleService) GetScheduleByGroup(ctx context.Context, group string, addEmptyLessons bool, dateStr string) (*models.StudentSchedule, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	startDate, endDate := utils.GetWeekBounds(date)

	group = strings.ToUpper(group)

	resp, err := s.Repo.GetScheduleByGroup(ctx, group, startDate, endDate)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedule for group %v not found", group)}
		}
		return nil, err
	}

	if addEmptyLessons {
		s.AddEmptyLessons(&resp.Schedule, resp.LessonsTimes)
	}
	return resp, err
}

func (s *ScheduleService) GetSchedulesByGroups(ctx context.Context, dateStr string, groups []string) ([]*models.StudentSchedule, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	startDate, endDate := utils.GetWeekBounds(date)
	resp, err := s.Repo.GetSchedulesByGroups(ctx, startDate, endDate, groups)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedules for groups `%v` not found", groups)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetGroups(ctx context.Context, facultyName string, course int, dateStr string) (*models.CourseFacultyGroups, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	const monthsOffset = 6
	startDate, endDate := utils.GetDateRangeBounds(date, monthsOffset)

	resp, err := s.Repo.GetGroups(ctx, facultyName, course, startDate, endDate)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("groups for faculty %v and course %v not found", facultyName, course)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetFaculties(ctx context.Context) (*models.Faculties, error) {
	resp, err := s.Repo.GetFaculties(ctx)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{"faculties not found"}
		}
		return nil, err
	}
	return resp, err
}
func (s *ScheduleService) GetFacultyCourses(ctx context.Context, facultyName string, dateStr string) (*models.FacultyCourses, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	const monthsOffset = 6
	startDate, endDate := utils.GetDateRangeBounds(date, monthsOffset)

	resp, err := s.Repo.GetFacultyCourses(ctx, facultyName, startDate, endDate)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("courses for faculty %v not found", facultyName)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetCourseFaculties(ctx context.Context, course int, dateStr string) (*models.CourseFaculties, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	const monthsOffset = 6
	startDate, endDate := utils.GetDateRangeBounds(date, monthsOffset)

	resp, err := s.Repo.GetCourseFaculties(ctx, course, startDate, endDate)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("faculties for course %v not found", course)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetDay(ctx context.Context) (*models.Day, error) {
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
		Time:        now.Format("15:04"),
	}, nil
}

func (s *ScheduleService) GetTeacherSchedule(ctx context.Context, teacherID int, dateStr string) (*models.TeacherSchedule, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}
	startDate, endDate := utils.GetWeekBounds(date)
	resp, err := s.Repo.GetTeacherSchedule(ctx, teacherID, startDate, endDate)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedule for teacher '%v' not found", teacherID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetAllTeachers(ctx context.Context) (*models.TeachersList, error) {
	return s.Repo.GetAllTeachers(ctx)
}

func (s *ScheduleService) GetTeachersList(ctx context.Context, facultyID, departmentID int) (*models.TeachersList, error) {
	resp, err := s.Repo.GetTeachersList(ctx, facultyID, departmentID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{
				fmt.Sprintf(
					"teachers for faculty '%v' and department '%v' not found",
					facultyID, departmentID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetTeachersFaculties(ctx context.Context, departmentID int) ([]*models.Faculty, error) {
	resp, err := s.Repo.GetTeachersFaculties(ctx, departmentID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("faculties for department with id '%v' not found", departmentID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetTeachersDepartments(ctx context.Context, facultyID int) ([]*models.Department, error) {
	resp, err := s.Repo.GetTeachersDepartments(ctx, facultyID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("departments for faculty '%v' not found", facultyID)}
		}
		return nil, err
	}
	return resp, err
}

func addEmptyLessons(lessons []models.StudentLesson, times []string) []models.StudentLesson {
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
			lessons = append(lessons, models.StudentLesson{
				Time:               time,
				Lesson:             emptyLessonText,
				TeacherAuditoriums: []models.StudentTeacherAuditorium{},
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

func (s *ScheduleService) GetFacultiesWithCourses(ctx context.Context, dateStr string) (*models.FacultiesCourses, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	const monthsOffset = 6
	startDate, endDate := utils.GetDateRangeBounds(date, monthsOffset)

	resp, err := s.Repo.GetFacultiesWithCourses(ctx, startDate, endDate)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{"no results"}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) AddEmptyLessons(schedule *models.NumeratorDenominator[models.StudentWeek], times []string) {
	schedule.Denominator.Monday = addEmptyLessons(schedule.Denominator.Monday, times)
	schedule.Denominator.Tuesday = addEmptyLessons(schedule.Denominator.Tuesday, times)
	schedule.Denominator.Wednesday = addEmptyLessons(schedule.Denominator.Wednesday, times)
	schedule.Denominator.Thursday = addEmptyLessons(schedule.Denominator.Thursday, times)
	schedule.Denominator.Friday = addEmptyLessons(schedule.Denominator.Friday, times)
	schedule.Denominator.Saturday = addEmptyLessons(schedule.Denominator.Saturday, times)

	schedule.Numerator.Monday = addEmptyLessons(schedule.Numerator.Monday, times)
	schedule.Numerator.Tuesday = addEmptyLessons(schedule.Numerator.Tuesday, times)
	schedule.Numerator.Wednesday = addEmptyLessons(schedule.Numerator.Wednesday, times)
	schedule.Numerator.Thursday = addEmptyLessons(schedule.Numerator.Thursday, times)
	schedule.Numerator.Friday = addEmptyLessons(schedule.Numerator.Friday, times)
	schedule.Numerator.Saturday = addEmptyLessons(schedule.Numerator.Saturday, times)
}

func (s *ScheduleService) GetAuditoriumSchedule(ctx context.Context, auditoriumID int, dateStr string) (*models.AuditoriumSchedule, error) {
	date, err := ParseDateOrNow(dateStr)
	if err != nil {
		return nil, err
	}

	startDate, endDate := utils.GetWeekBounds(date)
	resp, err := s.Repo.GetAuditoriumSchedule(ctx, startDate, endDate, auditoriumID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("schedules for auditorium `%v` not found", auditoriumID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetAuditorium(ctx context.Context, auditoriumID int) (*models.Auditorium, error) {
	resp, err := s.Repo.GetAuditorium(ctx, auditoriumID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("auditorium with id '%v' not found", auditoriumID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetAuditoriumsList(ctx context.Context, buildingID int) ([]*models.Auditorium, error) {
	resp, err := s.Repo.GetAuditoriumsList(ctx, buildingID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			if buildingID == 0 {
				return nil, NotFoundError{"auditoriums not found"}
			}
			return nil, NotFoundError{fmt.Sprintf("auditoriums for building '%v' not found", buildingID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetBuildingsList(ctx context.Context) ([]*models.Building, error) {
	resp, err := s.Repo.GetBuildingsList(ctx)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{"buildings not found"}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetBuilding(ctx context.Context, buildingID int) (*models.Building, error) {
	resp, err := s.Repo.GetBuilding(ctx, buildingID)
	if err != nil {
		if errors.Is(err, repo.ErrNoResults) {
			return nil, NotFoundError{fmt.Sprintf("building with id '%v' not found", buildingID)}
		}
		return nil, err
	}
	return resp, err
}

func (s *ScheduleService) GetLessonTypes() []models.LessonType {
	return []models.LessonType{
		{Type: "lecture", Description: "лекция"},
		{Type: "lab", Description: "лабораторная работа"},
		{Type: "practice", Description: "практика"},
		{Type: "coursework", Description: "курсовая работа"},
		{Type: "course_project", Description: "курсовой проект"},
		{Type: "exam", Description: "экзамен"},
		{Type: "zachet", Description: "зачет"},
		{Type: "consultation", Description: "консультация"},
		{Type: "elective", Description: "факультатив"},
		{Type: "unknown", Description: ""},
	}
}
