package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/schedule-rsreu/schedule-api/pkg/logger"

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

	scheduleGroup.GET("/day", sh.getDay) // /day

	scheduleGroup.GET("/courses", sh.getFacultyCourses) // /courses?faculty=фвт

	scheduleGroup.GET("/groups/:group", sh.getScheduleByGroup) // /groups/344
	scheduleGroup.POST("/groups/sample", sh.schedulesByGroups) // groups/sample
	scheduleGroup.GET("/groups", sh.getCourseFacultyGroups)    // /groups?faculty=фвт&course=3

	scheduleGroup.GET("/faculties", sh.getFaculties)                // /faculties
	scheduleGroup.GET("/faculties/course", sh.getCourseFaculties)   // /faculties/course?course=1
	scheduleGroup.GET("/faculties/courses", sh.getFacultiesCourses) // /faculties/course?course=1

	scheduleGroup.GET("/teachers", sh.getTeacherSchedule)                 // /teachers?teacher=Конюхов+Алексей+Николаевич
	scheduleGroup.GET("/teachers/all", sh.getTeachers)                    // /teachers/all
	scheduleGroup.GET("/teachers/list", sh.getTeachersList)               // /teachers/list?faculty=фаиту&department=ВМ
	scheduleGroup.GET("/teachers/departments", sh.getTeachersDepartments) // /teachers/departments?faculty=фаиту
	scheduleGroup.GET("/teachers/faculties", sh.getTeachersFaculties)     // /teachers/faculties?department=ВМ

	scheduleGroup.GET("/auditoriums", sh.getAuditoriumSchedule) // /auditoriums
	scheduleGroup.GET("/auditoriums/list", sh.getAuditoriumList)
	scheduleGroup.GET("/auditoriums/:auditorium_id", sh.getAuditorium)

	scheduleGroup.GET("/buildings", sh.getBuildings)
	scheduleGroup.GET("/buildings/:id", sh.getBuilding)

	scheduleGroup.GET("/lesson/types", sh.getLessonTypes) // /auditoriums
}

// getScheduleByGroup
// @Summary     Get schedule by group
// @Description Get schedule by group
// @Tags        Groups
// @Router      /api/v1/schedule/groups/{group} [get]
// @Param       group  path  string  true  "group" example(344)
// @Param       add_empty_lessons  query  bool  false  "add empty lessons"
// @Param       date  query  string  false  "date" example(2025-07-13)
// @Success     200  {object}  models.StudentSchedule
// @Response    200  {object}  models.StudentSchedule
// @Failure     500  {object}  echo.HTTPError.
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getScheduleByGroup(c echo.Context) error {
	ctx := c.Request().Context()
	group := c.Param("group")
	addEmptyLessons := c.QueryParam("add_empty_lessons") == "true"
	date := c.QueryParam("date")

	if group == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "group query param not found")
	}

	resp, err := sh.s.GetScheduleByGroup(ctx, group, addEmptyLessons, date)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	resp.LessonsTimes = nil

	return c.JSON(http.StatusOK, resp)
}

// getTeacherSchedule
// @Summary     Get teacher schedule
// @Description Расписание преподавателя
// @Tags        Teachers
// @Router      /api/v1/schedule/teachers [get]
// @Param       teacher_id  query  int  true  "teacher" example("Конюхов Алексей Николаевич")
// @Param       date  query  string  false  "date" example(2025-07-13)
// @Success     200  {object}  models.TeacherSchedule
// @Response    200  {object}  models.TeacherSchedule
// @Failure     500  {object}  echo.HTTPError.
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getTeacherSchedule(c echo.Context) error {
	teacherID := c.QueryParam("teacher_id")
	date := c.QueryParam("date")

	if teacherID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "teacher_id query param not found")
	}

	teacherIdInt, err := strconv.Atoi(teacherID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "teacher_id query param must be integer")
	}
	ctx := c.Request().Context()
	resp, err := sh.s.GetTeacherSchedule(ctx, teacherIdInt, date)

	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	resp.LessonsTimes = nil

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
	ctx := c.Request().Context()

	resp, err := sh.s.GetAllTeachers(ctx)
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
	ctx := c.Request().Context()

	resp, err := sh.s.GetDay(ctx)
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
	ctx := c.Request().Context()

	resp, err := sh.s.GetFaculties(ctx)
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
// @Description Курсы факультета. Фильтрует по наличию занятий в диапазоне ±6 месяцев от date. По умолчанию date = текущий день
// @Tags        Courses
// @Router      /api/v1/schedule/courses [get]
// @Param       faculty  query  string  true  "faculty" Enums(иэф, фаиту, фвт, фрт, фэ)
// @Param       date  query  string  false  "date" example(2025-01-08)
// @Success     200  {object}  models.FacultyCourses
// @Response    200  {object}  models.FacultyCourses
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getFacultyCourses(c echo.Context) error {
	faculty := c.QueryParam("faculty")
	if faculty == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "faculty query param not found")
	}

	date := c.QueryParam("date")
	ctx := c.Request().Context()

	resp, err := sh.s.GetFacultyCourses(ctx, faculty, date)
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
// @Description Факультеты курса. Фильтрует по наличию занятий в диапазоне ±6 месяцев от date. По умолчанию date = текущий день
// @Tags        Faculties
// @Router      /api/v1/schedule/faculties/course [get]
// @Param       course  query  int  true  "course" Enums(1, 2, 3, 4, 5, 6)
// @Param       date  query  string  false  "date" example(2025-01-08)
// @Success     200  {object}  models.CourseFaculties
// @Response    200  {object}  models.CourseFaculties
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getCourseFaculties(c echo.Context) error {
	course, err := strconv.Atoi(c.QueryParam("course"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "course query param must be integer")
	}

	date := c.QueryParam("date")
	ctx := c.Request().Context()

	resp, err := sh.s.GetCourseFaculties(ctx, course, date)
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

// getFacultiesCourses
// @Summary     Get faculties with courses
// @Description Факультеты с курсами. Фильтрует по наличию занятий в диапазоне ±6 месяцев от date. По умолчанию date = текущий день
// @Tags        Faculties
// @Router      /api/v1/schedule/faculties/courses [get]
// @Param       date  query  string  false  "date" example(2025-01-08)
// @Success     200  {object}  models.FacultiesCourses
// @Response    200  {object}  models.FacultiesCourses
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getFacultiesCourses(c echo.Context) error {
	date := c.QueryParam("date")
	ctx := c.Request().Context()

	resp, err := sh.s.GetFacultiesWithCourses(ctx, date)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// schedulesByGroups
// @Summary     Get schedules by groups
// @Description Рассписание для нескольких групп
// @Tags        Groups
// @Router      /api/v1/schedule/groups/sample [post]
// @Param       groups  body   schedulesByGroupsRequest  true  "groups"
// @Param       date  query  string  false  "date" example(2025-07-13)
// @Success     200  {array}   models.StudentSchedule
// @Response    200  {array}   models.StudentSchedule
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) schedulesByGroups(c echo.Context) error {
	var req schedulesByGroupsRequest

	date := c.QueryParam("date")

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	resp, err := sh.s.GetSchedulesByGroups(ctx, date, req.Groups)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	for _, schedule := range resp {
		schedule.LessonsTimes = nil
	}

	return c.JSON(http.StatusOK, resp)
}

// GetCourseFacultyGroups
// @Summary     Get course faculty groups
// @Description Группы факультета курса. Фильтрует по наличию занятий в диапазоне ±6 месяцев от date. По умолчанию date = текущий день
// @Tags        Groups
// @Router      /api/v1/schedule/groups [get]
// @Param       course  query  int  true  "course" Enums(1, 2, 3, 4, 5, 6)
// @Param       faculty  query  string  true  "faculty" Enums(иэф, фаиту, фвт, фрт, фэ)
// @Param       date  query  string  false  "date" example(2025-01-08)
// @Success     200  {array}   models.CourseFacultyGroups
// @Response    200  {array}   models.CourseFacultyGroups
// @Failure     400  {object}  echo.HTTPError
// @Failure     500  {object}  echo.HTTPError
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getCourseFacultyGroups(c echo.Context) error {
	faculty := c.QueryParam("faculty")
	if faculty == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "faculty query param is required")
	}
	faculty = strings.ToLower(faculty)

	courseS := c.QueryParam("course")
	if courseS == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "course query param is required")
	}

	course, err := strconv.Atoi(courseS)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "course query param must be integer, got: "+courseS)
	}

	date := c.QueryParam("date")

	ctx := c.Request().Context()

	resp, err := sh.s.GetGroups(ctx, faculty, course, date)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// GetTeachersList
// @Summary     Get teachers list by faculty and department
// @Description Список преподавателей по факультету и кафедре. Параметры не обязательны.
// @Tags        Teachers
// @Router      /api/v1/schedule/teachers/list [get]
// @Param       faculty_id  query  int  false  "faculty_id" example(4)
// @Param       department_id  query  int  false  "department_id" example(17)
// @Success     200  {object}   models.TeachersList
// @Response    200  {object}   models.TeachersList
// @Failure     500  {object}   echo.HTTPError
// @Failure     404  {object}   echo.HTTPError.
func (sh *ScheduleHandler) getTeachersList(c echo.Context) error {
	facultyID := c.QueryParam("faculty_id")
	departmentID := c.QueryParam("department_id")

	facultyIDInt, err := strconv.Atoi(facultyID)
	if facultyID != "" && err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "faculty_id query param must be integer, got: "+facultyID)
	}

	departmentIDInt, err := strconv.Atoi(departmentID)
	if departmentID != "" && err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "department_id query param must be integer, got: "+departmentID)
	}

	ctx := c.Request().Context()

	resp, err := sh.s.GetTeachersList(ctx, facultyIDInt, departmentIDInt)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		logger.GetLoggerFromCtx(c).Err(err).Msg("failed to get teachers list")
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getTeachersFaculties
// @Summary     Get faculties list by department
// @Description Список факультетов. Если кафедра не передан, то возвращаются все факультеты.
// @Tags        Teachers
// @Router      /api/v1/schedule/teachers/faculties [get]
// @Param       department_id  query  int  false  "department_id" example(123)
// @Success     200  {array}    models.Faculty
// @Response    200  {array}    models.Faculty
// @Failure     500  {object}   echo.HTTPError
// @Failure     404  {object}   echo.HTTPError.
func (sh *ScheduleHandler) getTeachersFaculties(c echo.Context) error {
	departmentId := c.QueryParam("department_id")

	departmentIdInt, err := strconv.Atoi(departmentId)
	if departmentId != "" && err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "department_id query param must be integer, got: "+departmentId)
	}

	ctx := c.Request().Context()

	resp, err := sh.s.GetTeachersFaculties(ctx, departmentIdInt)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getTeachersDepartments
// @Summary     Get departments list by faculty
// @Description Список кафедр. Если факультет не передана, то возвращаются все кафедры.
// @Tags        Teachers
// @Router      /api/v1/schedule/teachers/departments [get]
// @Param       faculty_id  query  int  false  "faculty_id" example(1)
// @Success     200  {array}    models.Department
// @Response    200  {array}    models.Department
// @Failure     500  {object}   echo.HTTPError
// @Failure     404  {object}   echo.HTTPError.
func (sh *ScheduleHandler) getTeachersDepartments(c echo.Context) error {
	facultyID := c.QueryParam("faculty_id")

	facultyIDInt, err := strconv.Atoi(facultyID)
	if facultyID != "" && err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "faculty_id query param must be integer, got: "+facultyID)
	}

	ctx := c.Request().Context()

	resp, err := sh.s.GetTeachersDepartments(ctx, facultyIDInt)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// getAuditoriumSchedule
// @Summary     Get auditorium schedule
// @Description Get auditorium schedule by auditorium_id
// @Tags        Auditoriums
// @Router      /api/v1/schedule/auditoriums [get]
// @Param       auditorium_id  query  int  true  "auditorium_id" example(12)
// @Param       date  query  string  false  "date" example(2025-06-13)
// @Success     200  {object}  models.AuditoriumSchedule
// @Response    200  {object}  models.AuditoriumSchedule
// @Failure     500  {object}  echo.HTTPError.
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getAuditoriumSchedule(c echo.Context) error {
	auditoriumIdStr := c.QueryParam("auditorium_id")

	if auditoriumIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "auditorium_id query param not found")
	}

	auditoriumIdInt, err := strconv.Atoi(auditoriumIdStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auditorium_id query param must be integer")
	}

	date := c.QueryParam("date")

	ctx := c.Request().Context()

	resp, err := sh.s.GetAuditoriumSchedule(ctx, auditoriumIdInt, date)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getAuditoriumList(c echo.Context) error {
	buildingIdStr := c.QueryParam("building_id")

	buildingIdInt := 0 // Default to 0 to get all auditoriums
	if buildingIdStr != "" {
		var err error
		buildingIdInt, err = strconv.Atoi(buildingIdStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "building_id query param must be integer")
		}
	}

	ctx := c.Request().Context()

	resp, err := sh.s.GetAuditoriumsList(ctx, buildingIdInt)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getAuditorium(c echo.Context) error {
	auditoriumIdStr := c.Param("auditorium_id")

	if auditoriumIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "auditorium_id path param not found")
	}

	auditoriumIdInt, err := strconv.Atoi(auditoriumIdStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auditorium_id path param must be integer")
	}

	ctx := c.Request().Context()

	resp, err := sh.s.GetAuditorium(ctx, auditoriumIdInt)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getBuildings(c echo.Context) error {
	ctx := c.Request().Context()

	resp, err := sh.s.GetBuildingsList(ctx)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getBuilding(c echo.Context) error {
	buildingIdStr := c.Param("id")

	if buildingIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id path param not found")
	}

	buildingIdInt, err := strconv.Atoi(buildingIdStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "id path param must be integer")
	}
	ctx := c.Request().Context()

	resp, err := sh.s.GetBuilding(ctx, buildingIdInt)
	if err != nil {
		if errors.As(err, &services.NotFoundError{}) {
			return echo.NewHTTPError(http.StatusNotFound, err)
		}
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// getLessonTypes
// @Summary     Get lesson types
// @Description Get lesson types
// @Tags        Lesson
// @Router      /api/v1/schedule/lesson/types [get]
// @Success     200  {array}  models.LessonType
// @Response    200  {object}  models.LessonType
// @Failure     500  {object}  echo.HTTPError.
// @Failure     404  {object}  echo.HTTPError.
func (sh *ScheduleHandler) getLessonTypes(c echo.Context) error {
	return c.JSON(http.StatusOK, sh.s.GetLessonTypes())
}
