package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewMongoClient(mongoURL string) (*mongo.Client, error) {
	const contextTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))

	if err != nil {
		return nil, fmt.Errorf("MongoDB connection failed: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %w", err)
	}

	return client, nil
}

func NewMongoDatabase(client *mongo.Client, dbName string) *mongo.Database {
	return client.Database(dbName)
}
