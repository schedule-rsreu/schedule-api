package repo

import (
	"context"
	"errors"
	"github.com/VinGP/schedule-api/scheme"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ScheduleRepo struct {
	mdb *mongo.Client
	c   *mongo.Collection
}

func New(mdb *mongo.Client) *ScheduleRepo {
	c := mdb.Database("schedule_database").Collection("schedule")
	return &ScheduleRepo{mdb, c}
}

func (sr *ScheduleRepo) GetScheduleByGroup(group string) (*scheme.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var schedule scheme.Schedule

	err := sr.c.FindOne(ctx, bson.M{"group": group}).Decode(&schedule)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNoScheduleGroup{group}
	} else if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (sr *ScheduleRepo) GetSchedulesByGroups(groups []string) ([]*scheme.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var schedules []*scheme.Schedule

	cursor, err := sr.c.Find(ctx, bson.M{"group": bson.M{"$in": groups}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (sr *ScheduleRepo) GetGroups(facultyName string, course int) (scheme.CourseFacultyGroups, error) {
	var stage mongo.Pipeline
	if facultyName == "" || course == 0 {
		stageBase := []bson.D{
			bson.D{{"$group", bson.D{
				{"_id", 0},
				{"group", bson.D{
					{"$addToSet", "$group"},
				}},
			}}},
			bson.D{{"$unwind", "$group"}},
			bson.D{{"$sort", bson.D{
				{"group", 1}},
			}},
			bson.D{{"$group", bson.D{
				{"_id", 0},
				{"_groups", bson.D{
					{"$push", "$group"}},
				},
			},
			}},
			bson.D{{"$project", bson.D{
				{"_id", 0},
				{"faculty", facultyName},
				{"course",
					bson.D{{"$literal", course}},
				},
				{"groups", "$_groups"},
			}}},
		}

		stageFacultyCourse := append(mongo.Pipeline{
			bson.D{{"$match", bson.D{
				{"faculty", facultyName},
				{"course", course},
			}}},
		}, stageBase...)

		//{
		//	"course",
		//		bson.D{{"$literal", course}},
		//},

		stageFaculty := append(mongo.Pipeline{
			bson.D{{"$match", bson.D{
				{"faculty", facultyName},
			}}},
		}, stageBase...)

		//stageCourse := append(mongo.Pipeline{
		//	bson.D{{"$match", {},
		//	}},
		//}, stageBase...)
		stage = stageBase
		if facultyName != "" && course != 0 {
			stage = stageFacultyCourse
		} else if facultyName != "" && course != 0 {
			stage = stageFacultyCourse
		} else if facultyName != "" {
			stage = stageFaculty
		} else {
			stage = append(mongo.Pipeline{
				bson.D{{"$match", bson.D{}}},
			}, stageBase...)
		}
	} else {
		stage = mongo.Pipeline{
			// Шаг $match
			bson.D{{
				"$match", bson.D{{
					"$expr", bson.D{{
						"$and", bson.A{
							// Фильтрация по faculty
							bson.D{{
								"$or", bson.A{
									bson.D{{"$eq", bson.A{"$faculty", "фвт"}}},
									bson.D{{"$eq", bson.A{"фвт", ""}}},
								},
							}},
							// Фильтрация по course
							bson.D{{
								"$or", bson.A{
									bson.D{{"$eq", bson.A{"$course", 1}}},
									bson.D{{"$eq", bson.A{1, 0}}},
								},
							}},
						},
					}},
				}},
			}},
			// Шаг $addFields
			bson.D{{
				"$addFields", bson.D{{
					"endsWithM", bson.D{{
						"$regexMatch", bson.D{{
							"input", "$group",
						}, {
							"regex", "м$",
						}},
					}},
				}},
			}},
			// Шаг $sort
			bson.D{{
				"$sort", bson.D{{
					"endsWithM", 1,
				}, {
					"group", 1,
				}},
			}},
			// Шаг $group
			bson.D{{
				"$group", bson.D{{"_id", bson.D{{"faculty", bson.D{{"$ifNull", bson.A{"$faculty", ""}}}}, {
					"course", bson.D{{"$ifNull", bson.A{"$course", 0}}},
				}},
				}, {"groups", bson.D{{"$push", "$group"}}}},
			}},

			// Шаг $project
			bson.D{{
				"$project", bson.D{{
					"_id", 0,
				}, {
					"faculty", "$_id.faculty",
				}, {
					"course", "$_id.course",
				}, {
					"groups", 1,
				}},
			}},
		}
	}

	cursor, err := sr.c.Aggregate(context.TODO(), stage)
	if err != nil {
		return scheme.CourseFacultyGroups{}, err
	}
	var results []scheme.CourseFacultyGroups
	cursor.Next(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		return scheme.CourseFacultyGroups{}, err
	}

	for _, result := range results {
		return result, nil
	}
	return scheme.CourseFacultyGroups{}, ErrNoResults
}

func (sr *ScheduleRepo) GetCourseFacultyGroups(facultyName string, course int) (scheme.CourseFacultyGroups, error) {

	stage := mongo.Pipeline{
		bson.D{{"$match", bson.D{
			{"faculty", facultyName},
			{"course", course},
		}}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"group", bson.D{
				{"$addToSet", "$group"},
			}},
		}}},
		bson.D{{"$unwind", "$group"}},
		bson.D{{"$sort", bson.D{
			{"group", 1}},
		}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"_groups", bson.D{
				{"$push", "$group"}},
			},
		},
		}},
		bson.D{{"$project", bson.D{
			{"_id", 0},
			{"faculty", facultyName},
			{"course",
				bson.D{{"$literal", course}},
			},
			{"groups", "$_groups"},
		}}},
	}
	cursor, err := sr.c.Aggregate(context.TODO(), stage)
	if err != nil {
		return scheme.CourseFacultyGroups{}, err
	}
	var results []scheme.CourseFacultyGroups
	cursor.Next(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		return scheme.CourseFacultyGroups{}, err
	}

	for _, result := range results {
		return result, nil
	}
	return scheme.CourseFacultyGroups{}, ErrNoResults
}

func (sr *ScheduleRepo) GetFacultyGroups(facultyName string) (scheme.CourseFacultyGroups, error) {
	stage := mongo.Pipeline{
		bson.D{{"$match", bson.D{
			{"faculty", facultyName},
		}}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"group", bson.D{
				{"$addToSet", "$group"},
			}},
		}}},
		bson.D{{"$unwind", "$group"}},
		bson.D{{"$sort", bson.D{
			{"group", 1}},
		}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"_groups", bson.D{
				{"$push", "$group"}},
			},
		},
		}},
		bson.D{{"$project", bson.D{
			{"_id", 0},
			{"faculty", facultyName},
			{"groups", "$_groups"},
		}}},
	}
	cursor, err := sr.c.Aggregate(context.TODO(), stage)
	if err != nil {
		return scheme.CourseFacultyGroups{}, err
	}
	var results []scheme.CourseFacultyGroups
	cursor.Next(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		return scheme.CourseFacultyGroups{}, err
	}

	for _, result := range results {
		return result, nil
	}
	return scheme.CourseFacultyGroups{}, ErrNoResults
}

func (sr *ScheduleRepo) GetFaculties() (scheme.Faculties, error) {
	stage := mongo.Pipeline{
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"group", bson.D{
				{"$addToSet", "$faculty"}},
			},
		},
		}},
		bson.D{{"$unwind", "$group"}},
		bson.D{{"$sort", bson.D{{"group", 1}}}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"faculties", bson.D{
				{"$push", "$group"}},
			},
		},
		}},
		bson.D{{"$project", bson.D{
			{"_id", 0},
			{"faculties", 1},
		}}},
	}
	cursor, err := sr.c.Aggregate(context.TODO(), stage)
	if err != nil {
		return scheme.Faculties{}, err
	}
	var results []scheme.Faculties
	cursor.Next(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		return scheme.Faculties{}, err
	}

	for _, result := range results {
		return result, nil
	}
	return scheme.Faculties{}, ErrNoResults
}

func (sr *ScheduleRepo) GetFacultyCourses(facultyName string) (scheme.FacultyCourses, error) {
	stage := mongo.Pipeline{
		bson.D{{"$match", bson.D{
			{"faculty", facultyName},
		}}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"group", bson.D{
				{"$addToSet", "$course"},
			}},
		}}},
		bson.D{{"$unwind", "$group"}},
		bson.D{{"$sort", bson.D{
			{"group", 1}},
		}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"_courses", bson.D{
				{"$push", "$group"}},
			},
		},
		}},
		bson.D{{"$project", bson.D{
			{"_id", 0},
			{"faculty", facultyName},
			{"courses", "$_courses"},
		}}},
	}
	cursor, err := sr.c.Aggregate(context.TODO(), stage)
	if err != nil {
		return scheme.FacultyCourses{}, err
	}
	var results []scheme.FacultyCourses
	cursor.Next(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		return scheme.FacultyCourses{}, err
	}

	for _, result := range results {
		return result, nil
	}
	return scheme.FacultyCourses{}, ErrNoResults
}

func (sr *ScheduleRepo) GetCourseFaculties(course int) (scheme.CourseFaculties, error) {
	stage := mongo.Pipeline{
		bson.D{{"$match", bson.D{
			{"course", course},
		}}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"group", bson.D{
				{"$addToSet", "$faculty"},
			}},
		}}},
		bson.D{{"$unwind", "$group"}},
		bson.D{{"$sort", bson.D{
			{"group", 1}},
		}},
		bson.D{{"$group", bson.D{
			{"_id", 0},
			{"_faculties", bson.D{
				{"$push", "$group"}},
			},
		},
		}},
		bson.D{{"$project", bson.D{
			{"_id", 0},
			{"course",
				bson.D{{"$literal", course}},
			},
			{"faculties", "$_faculties"},
		}}},
	}
	cursor, err := sr.c.Aggregate(context.TODO(), stage)
	if err != nil {
		return scheme.CourseFaculties{}, err
	}
	var results []scheme.CourseFaculties
	cursor.Next(context.TODO())
	if err = cursor.All(context.TODO(), &results); err != nil {
		return scheme.CourseFaculties{}, err
	}

	for _, result := range results {
		return result, nil
	}
	return scheme.CourseFaculties{}, ErrNoResults
}
