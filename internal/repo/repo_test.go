package repo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/schedule-rsreu/schedule-api/internal/repo"

	"github.com/schedule-rsreu/schedule-api/pkg/mongodb"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type TestDatabase struct {
	container  testcontainers.Container
	DbInstance *mongo.Database
	DbAddress  string
}

func SetupTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	container, dbInstance, dbAddr, err := createMongoContainer(ctx)
	cancel()
	if err != nil {
		log.Fatal("failed to setup test", err)
	}
	return &TestDatabase{
		container:  container,
		DbInstance: dbInstance,
		DbAddress:  dbAddr,
	}
}

func (tdb *TestDatabase) TearDown() {
	err := tdb.container.Terminate(context.Background())
	if err != nil {
		log.Fatal("failed to teardown test", err)
	}
}

func createMongoContainer(ctx context.Context) (testcontainers.Container, *mongo.Database, string, error) {
	var env = map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": "root",
		"MONGO_INITDB_ROOT_PASSWORD": "pass",
		"MONGO_INITDB_DATABASE":      "testdb",
	}
	var port = "27017/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo:6.0.9",
			ExposedPorts: []string{port},
			Env:          env,
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start container: %w", err)
	}

	p, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to get container external port: %w", err)
	}

	log.Println("mongo container ready and running at port: ", p.Port())

	uri := "mongodb://root:pass@localhost:" + p.Port()
	client, err := mongodb.NewMongoClient(uri)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to establish database connection: %w", err)
	}

	db := mongodb.NewMongoDatabase(client, "testdb")

	return container, db, uri, nil
}

type RepositorySuite struct {
	suite.Suite
	repository   *repo.ScheduleRepo
	testDatabase *TestDatabase
}

func (suite *RepositorySuite) SetupSuite() {
	suite.testDatabase = SetupTestDatabase()
	suite.repository = repo.NewScheduleRepo(suite.testDatabase.DbInstance)
}

func (suite *RepositorySuite) TearDownSuite() {
	err := suite.testDatabase.container.Terminate(context.Background())
	if err != nil {
		suite.T().Fatal(err)
		return
	}
}

func (suite *RepositorySuite) TestGetScheduleByGroup() {
	const testGroupDoc = `{
	  "_id": "67785bb8697ca3c668c109d0",
	  "group": "344",
	  "course": 2,
	  "faculty": "фвт",
	  "file": "https://rsreu.ru/component/docman/doc_download/19690-fvt-2-kurs-smena",
	  "file_hash": "2c12d4aad753f89d677daa1c9ca18e64",
	  "schedule": {
		"numerator": {
		  "monday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "л.Объектно-ориентированные языки и системы программирования\nст. преп.Васин А.О.   106а C",
			  "teachers_short": [
				"Васин А.О."
			  ],
			  "teachers_full": [
				"Васин Александр Олегович"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106а C"
			  ]
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "пр.Объектно-ориентированные языки и системы программирования\nст. преп.Васин А.О.   206-4 C",
			  "teachers_short": [
				"Васин А.О."
			  ],
			  "teachers_full": [
				"Васин Александр Олегович"
			  ],
			  "dates": [],
			  "auditoriums": [
				"206-4 C"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "лаб.Базы данных\nас.Федосова Е.Б.   122 С",
			  "teachers_short": [
				"Федосова Е.Б."
			  ],
			  "teachers_full": [
				"Федосова Елена Борисовна"
			  ],
			  "dates": [],
			  "auditoriums": []
			}
		  ],
		  "tuesday": [
			{
			  "time": "09.55-11.30",
			  "lesson": "пр.Иностранный язык, Иностранный язык\nст. преп. ИнЯзЗаволокин А.И.   316 C\n` +
		`ст. преп. ИнЯзТермышева Е.Н.   322 C",
			  "teachers_short": [
				"Заволокин А.И.",
				"Термышева Е.Н."
			  ],
			  "teachers_full": [
				"Заволокин Александр Иванович",
				"Термышева Елена Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"316 C",
				"322 C"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Учебная практика/Учебная практика\nдоц.Соколова Ю.С.   106 C",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106 C"
			  ]
			},
			{
			  "time": "13.35-15.10",
			  "lesson": "пр.Учебная практика/Учебная практика\nдоц.Соколова Ю.С.   106 C",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106 C"
			  ]
			}
		  ],
		  "wednesday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "л.Теория вероятностей и математическая статистика\nдоц.Соколова Ю.С.   358 C",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"358 C"
			  ]
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "л.Высшая математика\nдоц.Конюхов А.Н.   405 С",
			  "teachers_short": [
				"Конюхов А.Н."
			  ],
			  "teachers_full": [
				"Конюхов Алексей Николаевич"
			  ],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Теория вероятностей и математическая статистика\nдоц.Соколова Ю.С.   106 C",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106 C"
			  ]
			}
		  ],
		  "thursday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "со смены пр. Базы данных ас. Федосова Е. Б. 210 С",
			  "teachers_short": [
				"Федосова Е.Б."
			  ],
			  "teachers_full": [
				"Федосова Елена Борисовна"
			  ],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "л.Бухгалтерский учет\nст. преп.Горшкова Г.Н.   304 L",
			  "teachers_short": [
				"Горшкова Г.Н."
			  ],
			  "teachers_full": [
				"Горшкова Галина Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"304 L"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Бухгалтерский учет\nст. преп.Горшкова Г.Н.   304 L",
			  "teachers_short": [
				"Горшкова Г.Н."
			  ],
			  "teachers_full": [
				"Горшкова Галина Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"304 L"
			  ]
			},
			{
			  "time": "13.35-15.10",
			  "lesson": "л.Математическая логика и теория алгоритмов\nдоц.Проказникова Е.Н.   206-1 C",
			  "teachers_short": [
				"Проказникова Е.Н."
			  ],
			  "teachers_full": [
				"Проказникова Елена Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"206-1 C"
			  ]
			}
		  ],
		  "friday": [
			{
			  "time": "09.55-11.30",
			  "lesson": "л.Экономика промышленности и управление предприятием\nст. преп.Кутузова И.В.   448 C",
			  "teachers_short": [
				"Кутузова И.В."
			  ],
			  "teachers_full": [
				"Кутузова Ирина Васильевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"448 C"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Элективные дисциплины по физической культуре и спорту\n   Стадион РГРТУ C",
			  "teachers_short": [],
			  "teachers_full": [],
			  "dates": [],
			  "auditoriums": []
			}
		  ],
		  "saturday": []
		},
		"denominator": {
		  "monday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "л.Объектно-ориентированные языки и системы программирования\nст. преп.Васин А.О.   106а C",
			  "teachers_short": [
				"Васин А.О."
			  ],
			  "teachers_full": [
				"Васин Александр Олегович"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106а C"
			  ]
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "лаб.Объектно-ориентированные языки и системы программирования\n` +
		`ст. преп.Васин А.О.   206-3 C\nас.Щенёва Ю.Б.",
			  "teachers_short": [
				"Васин А.О.",
				"Щенёва Ю.Б."
			  ],
			  "teachers_full": [
				"Васин Александр Олегович",
				"Щенёва Юлия Борисовна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"206-3 C"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Высшая математика\nдоц.Конюхов А.Н.   451 C",
			  "teachers_short": [
				"Конюхов А.Н."
			  ],
			  "teachers_full": [
				"Конюхов Алексей Николаевич"
			  ],
			  "dates": [],
			  "auditoriums": [
				"451 C"
			  ]
			}
		  ],
		  "tuesday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "пр.Элективные дисциплины по физической культуре и спорту\n   Стадион РГРТУ C",
			  "teachers_short": [],
			  "teachers_full": [],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "пр.Теория вероятностей и математическая статистика\nдоц.Соколова Ю.С.   106 C",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106 C"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "лаб.Базы данных\nас.Федосова Е.Б.   122 С",
			  "teachers_short": [
				"Федосова Е.Б."
			  ],
			  "teachers_full": [
				"Федосова Елена Борисовна"
			  ],
			  "dates": [],
			  "auditoriums": []
			}
		  ],
		  "wednesday": [
			{
			  "time": "09.55-11.30",
			  "lesson": "л.Высшая математика\nдоц.Конюхов А.Н.   405 С",
			  "teachers_short": [
				"Конюхов А.Н."
			  ],
			  "teachers_full": [
				"Конюхов Алексей Николаевич"
			  ],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Высшая математика\nдоц.Конюхов А.Н.   446 C",
			  "teachers_short": [
				"Конюхов А.Н."
			  ],
			  "teachers_full": [
				"Конюхов Алексей Николаевич"
			  ],
			  "dates": [],
			  "auditoriums": [
				"446 C"
			  ]
			}
		  ],
		  "thursday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "пр.Высшая математика\nдоц.Конюхов А.Н.   451 C",
			  "teachers_short": [
				"Конюхов А.Н."
			  ],
			  "teachers_full": [
				"Конюхов Алексей Николаевич"
			  ],
			  "dates": [],
			  "auditoriums": [
				"451 C"
			  ]
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "пр.Математическая логика и теория алгоритмов\nдоц.Проказникова Е.Н.   106 C",
			  "teachers_short": [
				"Проказникова Е.Н."
			  ],
			  "teachers_full": [
				"Проказникова Елена Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"106 C"
			  ]
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Иностранный язык, Иностранный язык\nст. преп. ИнЯзЗаволокин А.И.   426 C\n` +
		`ст. преп. ИнЯзТермышева Е.Н.   310 C",
			  "teachers_short": [
				"Заволокин А.И.",
				"Термышева Е.Н."
			  ],
			  "teachers_full": [
				"Заволокин Александр Иванович",
				"Термышева Елена Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"310 C",
				"426 C"
			  ]
			}
		  ],
		  "friday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "л.Экономика промышленности и управление предприятием\nст. преп.Кутузова И.В.   448 C",
			  "teachers_short": [
				"Кутузова И.В."
			  ],
			  "teachers_full": [
				"Кутузова Ирина Васильевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"448 C"
			  ]
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "л. Базы данных доц. Гринченко Н. Н. 324 С",
			  "teachers_short": [
				"Гринченко Н.Н."
			  ],
			  "teachers_full": [
				"Гринченко Наталья Николаевна"
			  ],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "11.40-13.15",
			  "lesson": "пр.Элективные дисциплины по физической культуре и спорту\n   Стадион РГРТУ C",
			  "teachers_short": [],
			  "teachers_full": [],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "13.35-15.10",
			  "lesson": "лаб.Экономика промышленности и управление предприятием\nст. преп.Кутузова И.В.   501 L\n` +
		`ст. преп.Орлов П.А.",
			  "teachers_short": [
				"Кутузова И.В.",
				"Орлов П.А."
			  ],
			  "teachers_full": [
				"Кутузова Ирина Васильевна",
				"Орлов Павел Алексеевич"
			  ],
			  "dates": [],
			  "auditoriums": [
				"501 L"
			  ]
			}
		  ],
		  "saturday": [
			{
			  "time": "08.10-09.45",
			  "lesson": "пр.Учебная практика/Учебная практика\nдоц.Соколова Ю.С.   46 С",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": []
			},
			{
			  "time": "09.55-11.30",
			  "lesson": "пр.Учебная практика/Учебная практика\nдоц.Соколова Ю.С.   46 C",
			  "teachers_short": [
				"Соколова Ю.С."
			  ],
			  "teachers_full": [
				"Соколова Юлия Сергеевна"
			  ],
			  "dates": [],
			  "auditoriums": [
				"46 C"
			  ]
			}
		  ]
		},
		"day_lessons_times": {
		  "monday": [
			"08.10-09.45",
			"09.55-11.30",
			"11.40-13.15",
			"13.35-15.10"
		  ],
		  "tuesday": [
			"08.10-09.45",
			"09.55-11.30",
			"11.40-13.15",
			"13.35-15.10"
		  ],
		  "wednesday": [
			"08.10-09.45",
			"09.55-11.30",
			"11.40-13.15",
			"13.35-15.10"
		  ],
		  "thursday": [
			"08.10-09.45",
			"09.55-11.30",
			"11.40-13.15",
			"13.35-15.10"
		  ],
		  "friday": [
			"08.10-09.45",
			"09.55-11.30",
			"11.40-13.15",
			"13.35-15.10"
		  ],
		  "saturday": [
			"08.10-09.45",
			"09.55-11.30",
			"11.40-13.15",
			"13.35-15.10"
		  ]
		}
	  },
	  "updated_at": "2025-01-10T23:39:05.035Z"
	}`
	var data map[string]interface{}

	err := json.Unmarshal([]byte(testGroupDoc), &data)
	suite.Require().NoError(err)

	result, err := suite.testDatabase.
		DbInstance.
		Collection(repo.ScheduleCollectionName).
		InsertOne(context.Background(), data)
	suite.Require().NoError(err)

	defer func() {
		suite.Require().NoError(suite.testDatabase.
			DbInstance.
			Collection(repo.ScheduleCollectionName).
			Drop(context.Background()))
	}()
	suite.Equal(result.InsertedID, data["_id"])

	createdSchedule, err := suite.repository.GetScheduleByGroup("344")

	suite.Require().NoError(err)
	suite.Equal("344", createdSchedule.Group)
}

func (suite *RepositorySuite) TestGetGroups() {
	const testGroupDocs = `
	[
		{	  "group": "23434",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "345",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "016",
			  "course": 4,
			  "faculty": "фэ"
		}
	]
	`
	var data []interface{}

	err := json.Unmarshal([]byte(testGroupDocs), &data)
	suite.Require().NoError(err)

	result, err := suite.
		testDatabase.
		DbInstance.
		Collection(repo.ScheduleCollectionName).
		InsertMany(context.Background(), data)
	suite.Require().NoError(err)
	defer func() {
		suite.Require().NoError(suite.testDatabase.
			DbInstance.
			Collection(repo.ScheduleCollectionName).
			Drop(context.Background()))
	}()

	suite.Len(result.InsertedIDs, 3)

	groups, err := suite.repository.GetGroups("фвт", 3)
	suite.Require().NoError(err)

	suite.Len(groups.Groups, 2)

	groups, err = suite.repository.GetGroups("фэ", 0)
	suite.Require().NoError(err)
	suite.Len(groups.Groups, 1)

	groups, err = suite.repository.GetGroups("", 0)
	suite.Require().NoError(err)
	suite.Len(groups.Groups, 3)
}

func (suite *RepositorySuite) TestGetFaculties() {
	const testGroupDocs = `
	[
		{	  "group": "23434",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "345",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "016",
			  "course": 4,
			  "faculty": "фэ"
		}
	]
	`
	var data []interface{}

	err := json.Unmarshal([]byte(testGroupDocs), &data)
	suite.Require().NoError(err)

	result, err := suite.
		testDatabase.
		DbInstance.
		Collection(repo.ScheduleCollectionName).
		InsertMany(context.Background(), data)
	suite.Require().NoError(err)
	defer func() {
		suite.Require().NoError(suite.testDatabase.
			DbInstance.
			Collection(repo.ScheduleCollectionName).
			Drop(context.Background()))
	}()

	suite.Len(result.InsertedIDs, 3)

	faculties, err := suite.repository.GetFaculties()
	suite.Require().NoError(err)

	suite.Len(faculties.Faculties, 2)
}
func (suite *RepositorySuite) TestGetCourseFaculties() {
	const testGroupDocs = `
	[
		{	  "group": "23434",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "345",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "016",
			  "course": 4,
			  "faculty": "фэ"
		}
	]
	`
	var data []interface{}

	err := json.Unmarshal([]byte(testGroupDocs), &data)
	suite.Require().NoError(err)

	result, err := suite.
		testDatabase.
		DbInstance.
		Collection(repo.ScheduleCollectionName).
		InsertMany(context.Background(), data)
	suite.Require().NoError(err)
	suite.Len(result.InsertedIDs, 3)
	defer func() {
		suite.Require().NoError(suite.testDatabase.
			DbInstance.
			Collection(repo.ScheduleCollectionName).
			Drop(context.Background()))
	}()

	courseFaculties, err := suite.repository.GetCourseFaculties(3)
	suite.Require().NoError(err)

	suite.Len(courseFaculties.Faculties, 1)

	courseFaculties, err = suite.repository.GetCourseFaculties(4)
	suite.Require().NoError(err)

	suite.Len(courseFaculties.Faculties, 1)

	courseFaculties, err = suite.repository.GetCourseFaculties(1)
	suite.Nil(courseFaculties)
	suite.Require().Equal(err, repo.ErrNoResults)
}

func (suite *RepositorySuite) TestGetFacultyCourses() {
	const testGroupDocs = `
	[
		{	  "group": "23434",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "345",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "016",
			  "course": 4,
			  "faculty": "фэ"
		}
	]
	`
	var data []interface{}

	err := json.Unmarshal([]byte(testGroupDocs), &data)
	suite.Require().NoError(err)

	result, err := suite.
		testDatabase.
		DbInstance.
		Collection(repo.ScheduleCollectionName).
		InsertMany(context.Background(), data)
	suite.Require().NoError(err)
	defer func() {
		suite.Require().NoError(suite.testDatabase.
			DbInstance.
			Collection(repo.ScheduleCollectionName).
			Drop(context.Background()))
	}()
	suite.Len(result.InsertedIDs, 3)

	facultyCourses, err := suite.repository.GetFacultyCourses("фвт")
	suite.Require().NoError(err)
	suite.Len(facultyCourses.Courses, 1)

	facultyCourses, err = suite.repository.GetFacultyCourses("abc")
	suite.Nil(facultyCourses)
	suite.Require().Equal(err, repo.ErrNoResults)
}

func (suite *RepositorySuite) TestGetCourseFacultyGroups() {
	const testGroupDocs = `
	[
		{	  "group": "23434",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "345",
			  "course": 3,
			  "faculty": "фвт"
		},
		{	  "group": "016",
			  "course": 4,
			  "faculty": "фэ"
		}
	]
	`
	var data []interface{}

	err := json.Unmarshal([]byte(testGroupDocs), &data)
	suite.Require().NoError(err)

	result, err := suite.
		testDatabase.
		DbInstance.
		Collection(repo.ScheduleCollectionName).
		InsertMany(context.Background(), data)
	suite.Require().NoError(err)
	defer func() {
		suite.Require().NoError(suite.testDatabase.
			DbInstance.
			Collection(repo.ScheduleCollectionName).
			Drop(context.Background()))
	}()
	suite.Len(result.InsertedIDs, 3)

	courseGroups, err := suite.repository.GetCourseFacultyGroups("фвт", 3)
	suite.Require().NoError(err)
	suite.Len(courseGroups.Groups, 2)

	courseGroups, err = suite.repository.GetCourseFacultyGroups("фвт", 4)
	suite.Nil(courseGroups)
	suite.Require().Equal(err, repo.ErrNoResults)
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
