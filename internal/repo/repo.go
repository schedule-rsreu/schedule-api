package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const queryTimeout = 10 * time.Second
const maxFacultyShortLen = 8
const maxDepartmentShortLen = 6

type ScheduleRepo struct {
	mdb                        *mongo.Database
	scheduleCollection         *mongo.Collection
	teachersScheduleCollection *mongo.Collection
}

const ScheduleCollectionName = "schedule"
const TeachersScheduleCollectionName = "teachers_schedule"

func NewScheduleRepo(mdb *mongo.Database) *ScheduleRepo {
	scheduleCollection := mdb.Collection(ScheduleCollectionName)
	teachersScheduleCollection := mdb.Collection(TeachersScheduleCollectionName)
	return &ScheduleRepo{mdb, scheduleCollection, teachersScheduleCollection}
}

func findOne[T any](filter any, c *mongo.Collection) (*T, error) {
	var result *T
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	err := c.FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNoResults
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func findAll[T any](filter any, c *mongo.Collection) ([]*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var results []*T

	cursor, err := c.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func aggregateOne[T any](pipeline any, c *mongo.Collection) (*T, error) {
	var results []*T

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	cursor.Next(context.TODO())
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		return result, nil
	}

	return nil, ErrNoResults
}

func aggregateAll[T any](pipeline any, c *mongo.Collection) ([]*T, error) {
	var results []*T

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrNoResults
	}

	return results, nil
}

func (sr *ScheduleRepo) GetScheduleByGroup(group string) (*models.Schedule, error) {
	return findOne[models.Schedule](bson.D{{Key: "group", Value: group}}, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetSchedulesByGroups(groups []string) ([]*models.Schedule, error) {
	return findAll[models.Schedule](bson.D{{Key: "group", Value: bson.M{"$in": groups}}}, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetGroups(facultyName string, course int) (*models.CourseFacultyGroups, error) { //nolint:funlen,lll // too long queries
	stageBase := []bson.D{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "group", Value: bson.D{
				{Key: "$addToSet", Value: "$group"},
			}},
		}}},
		{{Key: "$unwind", Value: "$group"}},
		{{Key: "$sort", Value: bson.D{
			{Key: "group", Value: 1}},
		}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "_groups", Value: bson.D{
				{Key: "$push", Value: "$group"}},
			},
		},
		}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculty", Value: facultyName},
			{Key: "course",
				Value: bson.D{{Key: "$literal", Value: course}},
			},
			{Key: "groups", Value: "$_groups"},
		}}},
	}

	var stage mongo.Pipeline

	if facultyName == "" || course == 0 {
		if facultyName != "" {
			stage = append(stage, mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "faculty", Value: facultyName},
				}}},
			}...)
		}

		if course != 0 {
			stage = append(stage, mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "course", Value: course},
				}}},
			}...)
		}

		stage = append(stage, stageBase...)
	} else {
		stage = mongo.Pipeline{
			bson.D{{
				Key: "$match", Value: bson.D{{
					Key: "$expr", Value: bson.D{{
						Key: "$and", Value: bson.A{
							bson.D{{
								Key: "$or", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$faculty", facultyName}}},
									bson.D{{Key: "$eq", Value: bson.A{facultyName, ""}}},
								},
							}},
							bson.D{{
								Key: "$or", Value: bson.A{
									bson.D{{Key: "$eq", Value: bson.A{"$course", course}}},
									bson.D{{Key: "$eq", Value: bson.A{course, 0}}},
								},
							}},
						},
					}},
				}},
			}},
			bson.D{{
				Key: "$addFields", Value: bson.D{{
					Key: "endsWithM", Value: bson.D{{
						Key: "$regexMatch", Value: bson.D{{
							Key: "input", Value: "$group",
						}, {
							Key: "regex", Value: "м$",
						}},
					}},
				}},
			}},
			bson.D{{
				Key: "$sort", Value: bson.D{{
					Key: "endsWithM", Value: 1,
				}, {
					Key: "group", Value: 1,
				}},
			}},
			bson.D{{
				Key: "$group",
				Value: bson.D{{Key: "_id", Value: bson.D{{Key: "faculty",
					Value: bson.D{{Key: "$ifNull",
						Value: bson.A{"$faculty", ""}}}}, {
					Key: "course", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$course", 0}}},
				}},
				}, {Key: "groups", Value: bson.D{{Key: "$push", Value: "$group"}}}},
			}},
			bson.D{{
				Key: "$project", Value: bson.D{{
					Key: "_id", Value: 0,
				}, {
					Key: "faculty", Value: facultyName,
				}, {
					Key: "course", Value: bson.D{{Key: "$literal", Value: course}},
				}, {
					Key: "groups", Value: 1,
				}},
			}},
		}
	}

	return aggregateOne[models.CourseFacultyGroups](stage, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetCourseFacultyGroups(facultyName string, course int) (*models.CourseFacultyGroups, error) {
	stage := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "faculty", Value: facultyName},
			{Key: "course", Value: course},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "group", Value: bson.D{
				{Key: "$addToSet", Value: "$group"},
			}},
		}}},
		bson.D{{Key: "$unwind", Value: "$group"}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "group", Value: 1}},
		}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "_groups", Value: bson.D{
				{Key: "$push", Value: "$group"}},
			},
		},
		}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculty", Value: facultyName},
			{Key: "course",
				Value: bson.D{{Key: "$literal", Value: course}},
			},
			{Key: "groups", Value: "$_groups"},
		}}},
	}

	return aggregateOne[models.CourseFacultyGroups](stage, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetFaculties() (*models.Faculties, error) {
	stage := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "group", Value: bson.D{
				{Key: "$addToSet", Value: "$faculty"}},
			},
		},
		}},
		bson.D{{Key: "$unwind", Value: "$group"}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "group", Value: 1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculties", Value: bson.D{
				{Key: "$push", Value: "$group"}},
			},
		},
		}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculties", Value: 1},
		}}},
	}

	return aggregateOne[models.Faculties](stage, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetFacultyCourses(facultyName string) (*models.FacultyCourses, error) {
	stage := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "faculty", Value: facultyName},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "group", Value: bson.D{
				{Key: "$addToSet", Value: "$course"},
			}},
		}}},
		bson.D{{Key: "$unwind", Value: "$group"}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "group", Value: 1}},
		}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "_courses", Value: bson.D{
				{Key: "$push", Value: "$group"}},
			},
		},
		}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculty", Value: facultyName},
			{Key: "courses", Value: "$_courses"},
		}}},
	}

	return aggregateOne[models.FacultyCourses](stage, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetFacultiesWithCourses() (*models.FacultiesCourses, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$faculty"},
			{Key: "courses", Value: bson.D{{Key: "$addToSet", Value: "$course"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculty", Value: "$_id"},
			{Key: "courses", Value: bson.D{{Key: "$setIntersection", Value: bson.A{"$courses", "$courses"}}}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "faculty", Value: 1},
		}}},
	}

	fc, err := aggregateAll[models.FacultyCourses](pipeline, sr.scheduleCollection)

	if err != nil {
		return nil, err
	}
	var res models.FacultiesCourses
	for _, f := range fc {
		res = append(res, *f)
	}

	return &res, nil
}

func (sr *ScheduleRepo) GetCourseFaculties(course int) (*models.CourseFaculties, error) {
	stage := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "course", Value: course},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "group", Value: bson.D{
				{Key: "$addToSet", Value: "$faculty"},
			}},
		}}},
		bson.D{{Key: "$unwind", Value: "$group"}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "group", Value: 1}},
		}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "_faculties", Value: bson.D{
				{Key: "$push", Value: "$group"}},
			},
		},
		}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "course",
				Value: bson.D{{Key: "$literal", Value: course}},
			},
			{Key: "faculties", Value: "$_faculties"},
		}}},
	}

	return aggregateOne[models.CourseFaculties](stage, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetTeacherSchedule(teacher string) (*models.TeacherSchedule, error) {
	filter := bson.D{{Key: "teacher_full", Value: teacher}}

	return findOne[models.TeacherSchedule](filter, sr.teachersScheduleCollection)
}

func (sr *ScheduleRepo) GetAllTeachers() (*models.TeachersList, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "teachers", Value: bson.D{{Key: "$addToSet", Value: "$teacher_full"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "teachers", Value: bson.D{{Key: "$sortArray", Value: bson.D{
				{Key: "input", Value: "$teachers"},
				{Key: "sortBy", Value: 1},
			}}}},
		}}},
	}

	return aggregateOne[models.TeachersList](pipeline, sr.teachersScheduleCollection)
}

func (sr *ScheduleRepo) GetTeachersFaculties(department *string) ([]*models.TeacherFaculty, error) {
	matchStage := bson.D{}

	if department != nil && *department != "" {
		if len(*department) > maxDepartmentShortLen {
			matchStage = append(matchStage, bson.E{Key: "department", Value: *department})
		} else {
			matchStage = append(matchStage, bson.E{Key: "department_short", Value: *department})
		}
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$match", Value: bson.D{
			{Key: "faculty", Value: bson.D{{Key: "$ne", Value: nil}}},
			{Key: "faculty_short", Value: bson.D{{Key: "$ne", Value: nil}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "faculty", Value: "$faculty"},
				{Key: "faculty_short", Value: "$faculty_short"},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "faculty", Value: "$_id.faculty"},
			{Key: "faculty_short", Value: "$_id.faculty_short"},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "faculty", Value: 1},
		}}},
	}

	return aggregateAll[models.TeacherFaculty](pipeline, sr.teachersScheduleCollection)
}

func (sr *ScheduleRepo) GetTeachersDepartments(faculty *string) ([]*models.TeacherDepartment, error) {
	matchStage := bson.D{}
	if faculty != nil && *faculty != "" {
		if len(*faculty) > maxFacultyShortLen {
			matchStage = append(matchStage, bson.E{Key: "faculty", Value: *faculty})
		} else {
			matchStage = append(matchStage, bson.E{Key: "faculty_short", Value: strings.ToLower(*faculty)})
		}
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$match", Value: bson.D{
			{Key: "department", Value: bson.D{{Key: "$ne", Value: nil}}},
			{Key: "department_short", Value: bson.D{{Key: "$ne", Value: nil}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "department", Value: "$department"},
				{Key: "department_short", Value: "$department_short"},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "department", Value: "$_id.department"},
			{Key: "department_short", Value: "$_id.department_short"},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "department", Value: 1},
		}}},
	}

	return aggregateAll[models.TeacherDepartment](pipeline, sr.teachersScheduleCollection)
}

func (sr *ScheduleRepo) GetTeachersList(faculty, department *string) (*models.TeachersList, error) {
	matchStage := bson.D{}
	if faculty != nil && *faculty != "" {
		if len(*faculty) > maxFacultyShortLen {
			matchStage = append(matchStage, bson.E{Key: "faculty", Value: strings.ToLower(*faculty)})
		} else {
			matchStage = append(matchStage, bson.E{Key: "faculty_short", Value: *faculty})
		}
	}
	if department != nil && *department != "" {
		if len(*department) > maxDepartmentShortLen {
			matchStage = append(matchStage, bson.E{Key: "department", Value: *department})
		} else {
			matchStage = append(matchStage, bson.E{Key: "department_short", Value: *department})
		}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "teachers", Value: bson.D{{Key: "$addToSet", Value: "$teacher_full"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "teachers", Value: bson.D{{Key: "$sortArray", Value: bson.D{
				{Key: "input", Value: "$teachers"},
				{Key: "sortBy", Value: 1},
			}}}},
		}}},
	}

	return aggregateOne[models.TeachersList](pipeline, sr.teachersScheduleCollection)
}
