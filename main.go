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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://Test:<password>@rgrty.dlo7r6h.mongodb.net/?retryWrites=true&w=majority&appName=RGRTY"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://Test:test@rgrty.dlo7r6h.mongodb.net/"))

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
