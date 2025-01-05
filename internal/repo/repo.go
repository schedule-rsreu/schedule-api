package repo

import (
	"context"
	"errors"
	"time"

	"github.com/schedule-rsreu/schedule-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const queryTimeout = 10 * time.Second

type ScheduleRepo struct {
	mdb                        *mongo.Client
	scheduleCollection         *mongo.Collection
	teachersScheduleCollection *mongo.Collection
}

func New(mdb *mongo.Client) *ScheduleRepo {
	scheduleCollection := mdb.Database("schedule_database").Collection("schedule")
	teachersScheduleCollection := mdb.Database("schedule_database").Collection("teachers_schedule")
	return &ScheduleRepo{mdb, scheduleCollection, teachersScheduleCollection}
}

// var ErrNoResults = errors.New("no results")

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

func (sr *ScheduleRepo) GetScheduleByGroup(group string) (*models.Schedule, error) {
	return findOne[models.Schedule](bson.D{{Key: "group", Value: group}}, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetSchedulesByGroups(groups []string) ([]*models.Schedule, error) {
	return findAll[models.Schedule](bson.D{{Key: "group", Value: bson.M{"$in": groups}}}, sr.scheduleCollection)
}

func (sr *ScheduleRepo) GetGroups(facultyName string, course int) (*models.CourseFacultyGroups, error) { //nolint:funlen,lll // too long queries
	stageBase := []bson.D{
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

	stageFacultyCourse := append(mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "faculty", Value: facultyName},
			{Key: "course", Value: course},
		}}},
	}, stageBase...)

	stageFaculty := append(mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "faculty", Value: facultyName},
		}}},
	}, stageBase...)

	var stage mongo.Pipeline

	if facultyName == "" || course == 0 {
		switch {
		case facultyName != "" && course != 0:
			stage = stageFacultyCourse
		case facultyName != "":
			stage = stageFaculty
		default:
			stage = append(mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{}}},
			}, stageBase...)
		}
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
