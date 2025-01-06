package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/schedule-rsreu/schedule-api/internal/services"

	"github.com/labstack/echo/v4"
	_ "github.com/schedule-rsreu/schedule-api/internal/models"
)

type ScheduleHandler struct {
	s *services.ScheduleService
}

func NewRouter(g *echo.Group,
	scheduleService *services.ScheduleService,

) {
	sh := &ScheduleHandler{
		s: scheduleService,
	}

	scheduleGroup := g.Group("/schedule")

	scheduleGroup.GET("/groups/:group", sh.getScheduleByGroup)    // /groups/344
	scheduleGroup.GET("/teachers", sh.getTeacherSchedule)         // /teachers?teacher=Конюхов+Алексей+Николаевич
	scheduleGroup.GET("/teachers/all", sh.getTeachers)            // /teachers/all
	scheduleGroup.GET("/day", sh.getDay)                          // /day
	scheduleGroup.GET("/faculties", sh.getFaculties)              // /faculties
	scheduleGroup.GET("/courses", sh.getFacultyCourses)           // /courses?faculty=фвт
	scheduleGroup.GET("/faculties/course", sh.getCourseFaculties) // /faculties/course?course=1
	scheduleGroup.POST("/groups/sample", sh.schedulesByGroups)    // groups/sample
	scheduleGroup.GET("/groups", sh.getCourseFacultyGroups)       // /groups?faculty=фвт&course=3
}

// getScheduleByGroup
// @Summary     Get schedule by group
// @Description Get schedule by group
// @Tags        Groups
// @Router      /api/v1/schedule/groups/{group} [get]
// @Param       group  path  string  true  "group" example(344)
// @Success     200  {object}  models.Schedule
// @Response    200  {object}  models.Schedule
// @Failure     500  {object}  echo.HTTPError.
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getScheduleByGroup(c echo.Context) error {
	group := c.Param("group")

	if group == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "group query param not found")
	}

	resp, err := sh.s.GetScheduleByGroup(group)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// getTeacherSchedule
// @Summary     Get teacher schedule
// @Description Расписание преподавателя
// @Tags        Teachers
// @Router      /api/v1/schedule/teachers [get]
// @Param       teacher  query  string  true  "teacher" example("Конюхов Алексей Николаевич")
// @Success     200  {object}  models.TeacherSchedule
// @Response    200  {object}  models.TeacherSchedule
// @Failure     500  {object}  echo.HTTPError.
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getTeacherSchedule(c echo.Context) error {
	teacher := c.QueryParam("teacher")

	if teacher == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "teacher query param not found")
	}

	resp, err := sh.s.GetTeacherSchedule(teacher)

	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// getTeachers
// @Summary     Get teachers
// @Description Список всех преподавателей
// @Tags        Teachers
// @Router      /api/v1/schedule/teachers/all [get]
// @Success     200  {object}  models.TeachersList
// @Response    200  {object}  models.TeachersList
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getTeachers(c echo.Context) error {
	resp, err := sh.s.GetTeachers()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getDay
// @Summary     Get day
// @Description Информация о текущем дне
// @Tags        Day
// @Router      /api/v1/schedule/day [get]
// @Success     200  {object}  models.Day
// @Response    200  {object}  models.Day
// @Failure     500  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getDay(c echo.Context) error {
	resp, err := sh.s.GetDay()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getFaculties
// @Summary     Get faculties
// @Description Факультеты
// @Tags        Faculties
// @Router      /api/v1/schedule/faculties [get]
// @Success     200  {object}  models.Faculties
// @Response    200  {object}  models.Faculties
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getFaculties(c echo.Context) error {
	resp, err := sh.s.GetFaculties()
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getFacultyCourses
// @Summary     Get faculty courses
// @Description Курсы факультета
// @Tags        Courses
// @Router      /api/v1/schedule/courses [get]
// @Param       faculty  query  string  true  "faculty" Enums(иэф, фаиту, фвт, фрт, фэ)
// @Success     200  {object}  models.FacultyCourses
// @Response    200  {object}  models.FacultyCourses
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getFacultyCourses(c echo.Context) error {
	faculty := c.QueryParam("faculty")
	if faculty == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "faculty query param not found")
	}

	resp, err := sh.s.GetFacultyCourses(faculty)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getCourseFaculties
// @Summary     Get course faculties
// @Description Факультеты курса
// @Tags        Faculties
// @Router      /api/v1/schedule/faculties/course [get]
// @Param       course  query  int  true  "course" Enums(1, 2, 3, 4, 5)
// @Success     200  {object}  models.CourseFaculties
// @Response    200  {object}  models.CourseFaculties
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getCourseFaculties(c echo.Context) error {
	course, err := strconv.Atoi(c.QueryParam("course"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "course query param must be integer")
	}

	resp, err := sh.s.GetCourseFaculties(course)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

type schedulesByGroupsRequest struct {
	Groups []string `json:"groups" validate:"required" example:"344,345,346"`
}

// schedulesByGroups
// @Summary     Get schedules by groups
// @Description Рассписание для нескольких групп
// @Tags        Groups
// @Router      /api/v1/schedule/groups/sample [post]
// @Param       groups  body   schedulesByGroupsRequest  true  "groups"
// @Success     200  {array}   models.Schedule
// @Response    200  {array}   models.Schedule
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) schedulesByGroups(c echo.Context) error {
	var req schedulesByGroupsRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := sh.s.GetSchedulesByGroups(req.Groups)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// GetCourseFacultyGroups
// @Summary     Get course faculty groups
// @Description Группы факультета курса
// @Tags        Groups
// @Router      /api/v1/schedule/groups [get]
// @Param       course  query  int  true  "course" Enums(1, 2, 3, 4, 5)
// @Param       faculty  query  string  true  "faculty" Enums(иэф, фаиту, фвт, фрт, фэ)
// @Success     200  {array}   models.CourseFacultyGroups
// @Response    200  {array}   models.CourseFacultyGroups
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getCourseFacultyGroups(c echo.Context) error {
	faculty := c.QueryParam("faculty")
	faculty = strings.ToLower(faculty)

	courseS := c.QueryParam("course")
	var course int
	var err error
	if courseS != "" {
		course, err = strconv.Atoi(courseS)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "course query param must be integer, got: "+courseS)
		}
	}

	resp, err := sh.s.GetGroups(faculty, course)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
