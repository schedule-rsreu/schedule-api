package v1

import (
	"errors"
	_ "github.com/VinGP/schedule-api/docs"
	"github.com/VinGP/schedule-api/repo"
	"github.com/VinGP/schedule-api/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type scheduleRoutes struct {
	s services.ScheduleService
}

func newScheduleRoutes(handler *gin.RouterGroup, s services.ScheduleService) {
	r := &scheduleRoutes{s}
	h := handler.Group("/schedule")
	{
		h.GET("/groups/:group", r.scheduleByGroup)       // /groups/344
		h.GET("/groups", r.getGroups)                    // /faculty/course?faculty=фвт&course=3
		h.GET("/groups/sample", r.schedulesByGroups)     // /groups
		h.GET("/faculties", r.getFaculties)              // /faculties
		h.GET("/course/faculties", r.getCourseFaculties) // /course/faculties?course=3
		h.GET("/courses", r.getFacultyCourses)           // /courses?faculty=фвт
		h.GET("/day", r.getDay)                          // /day
	}
}

/*
информация о группе (факультет, курс)
получения списка всех групп если не указаны факультет и курс
получение расписания по дню недели
получение расписания по дню недели и числителю/знаменателю
получения что сегодня - числитель/знаменатель
*/

// @Summary     Show schedule by group
// @Description Выдает расписание по группе
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.Schedule
// @Failure     500 {object} response
// @Param       group  path string  true  "search schedule by group" example(344)
// @Router       /schedule/groups/{group} [get]
func (r *scheduleRoutes) scheduleByGroup(c *gin.Context) {
	group := c.Param("group")
	schedule, err := r.s.GetScheduleByGroup(group)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, schedule)
}

type schedulesByGroupsRequest struct {
	Groups []string `json:"groups" example:"344,345,346" required:"true"`
}

// @Summary     Show schedules by groups
// @Description Выдает расписания по группам
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {array} scheme.Schedule
// @Failure     500 {object} response
// @Param       groups  body schedulesByGroupsRequest  true  "search schedules by groups"
// @Router       /schedule/groups/sample [get]
func (r *scheduleRoutes) schedulesByGroups(c *gin.Context) {
	var req schedulesByGroupsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	schedules, err := r.s.GetSchedulesByGroups(req.Groups)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// @Summary     Show groups by faculty and course
// @Description Выдает список групп на определенном курсе определенного факультета. Если курс не указан выдет все группы факультета
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.CourseFacultyGroups
// @Failure     500 {object} response
// @Param		faculty	query	string 	false	"факультет" Enums(иэф, фаиту, фвт, фрт, фэ)
// @Param		course	query	int 	false	"курс" 		Enums(1, 2, 3, 4, 5)
// @Router      /schedule/groups [get]
func (r *scheduleRoutes) getGroups(c *gin.Context) {
	faculty := c.Query("faculty")

	courseS := c.Query("course")
	var course int
	var err error
	if courseS != "" {
		course, err = strconv.Atoi(courseS)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "course must be integer")
			return
		}
	} else {
		course = 0
	}

	res, err := r.s.GetGroups(faculty, course)
	if errors.Is(err, repo.ErrNoResults) {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	} else if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)

}

// @Summary     Show faculties
// @Description Выдает список всех факультетов
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.Faculties
// @Failure     500 {object} response
// @Router      /schedule/faculties [get]
func (r *scheduleRoutes) getFaculties(c *gin.Context) {
	res, err := r.s.GetFaculties()
	if errors.Is(err, repo.ErrNoResults) {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	} else if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary     Show faculties by course
// @Description Выдает список факультетов определенного курса
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.CourseFaculties
// @Failure     500 {object} response
// @Param		course	query	int 	true	"курс" 		Enums(1, 2, 3, 4, 5)
// @Router      /schedule/course/faculties [get]
func (r *scheduleRoutes) getCourseFaculties(c *gin.Context) {
	courseS := c.Query("course")
	var course int
	var err error
	if courseS != "" {
		course, err = strconv.Atoi(courseS)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "course must be integer")
			return
		}
	} else {
		course = 0
	}
	res, err := r.s.GetCourseFaculties(course)
	if errors.Is(err, repo.ErrNoResults) {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	} else if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary     Show courses by faculty
// @Description Выдает список номеров курсов факультета
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.FacultyCourses
// @Failure     500 {object} response
// @Param		faculty	query 	string 	true	"факультет" Enums(иэф, фаиту, фвт, фрт, фэ)
// @Router      /schedule/courses [get]
func (r *scheduleRoutes) getFacultyCourses(c *gin.Context) {
	faculty := c.Query("faculty")
	if faculty == "" {
		errorResponse(c, http.StatusBadRequest, "param faculty dont exist")
		return
	}
	res, err := r.s.GetFacultyCourses(faculty)

	if errors.Is(err, repo.ErrNoResults) {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	} else if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary     Show day data
// @Description Текущий день
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.Day
// @Failure     500 {object} response
// @Router      /schedule/day [get]
func (r *scheduleRoutes) getDay(c *gin.Context) {

	res, err := r.s.GetDay()

	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}
