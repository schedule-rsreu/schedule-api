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
		h.GET("/groups/:group", r.scheduleByGroup) // /groups/344
		h.GET("/groups", r.getGroups)              // /faculty/course?faculty=фвт&course=3
		h.GET("/faculties", r.getFaculties)        // /faculties
		h.GET("/courses", r.getFacultyCourses)     // /courses?faculty=фвт
	}
}

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

// @Summary     Show groups by faculty and course
// @Description Выдает список групп на определенном курсе определенного факультета. Если курс не указан выдет все группы факультета
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.CourseFacultyGroups
// @Failure     500 {object} response
// @Param		faculty	query	string 	true	"факультет" example(фвт)
// @Param		course	query	int 	false	"курс" 		example(1)
// @Router      /schedule/groups [get]
func (r *scheduleRoutes) getGroups(c *gin.Context) {
	//faculty := c.Param("faculty")
	faculty := c.Query("faculty")
	if faculty == "" {
		errorResponse(c, http.StatusBadRequest, "param faculty dont exist")
		return
	}
	courseS := c.Query("course")

	if courseS == "" {
		res, err := r.s.GetFacultyGroups(faculty)
		if errors.Is(err, repo.ErrNoResults) {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		} else if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, res)
		return
	} else {
		course, err := strconv.Atoi(courseS)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "course must be integer")
			return
		}
		if course == 0 {
			errorResponse(c, http.StatusBadRequest, "param course dont exist")
			return
		}

		res, err := r.s.GetCourseFacultyGroups(faculty, course)
		if errors.Is(err, repo.ErrNoResults) {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		} else if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, res)
	}
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

// @Summary     Show courses by faculty
// @Description Выдает список номеров курсов факультета
// @Tags  	    schedule
// @Accept      json
// @Produce     json
// @Success     200 {object} scheme.FacultyCourses
// @Failure     500 {object} response
// @Param		faculty	query 	string 	true	"факультет" example(фвт)
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
