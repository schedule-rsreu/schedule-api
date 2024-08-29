package main

import (
	"context"
	v1 "github.com/VinGP/schedule-api/api/v1"
	"github.com/VinGP/schedule-api/repo"
	"github.com/VinGP/schedule-api/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:rootPassXXX@mongodb:27017/"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(CORSMiddleware())
	err = r.SetTrustedProxies(nil) //disabled Trusted Proxies
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	s := services.ScheduleService{Repo: repo.New(client)}
	v1.NewRouter(r, s)

	err = r.Run("0.0.0.0:8081")
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}
