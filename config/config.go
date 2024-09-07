package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"sync"
)

type Config struct {
	Port          string `env:"PORT" env-default:"80"`
	Host          string `env:"HOST" env-default:"0.0.0.0"`
	Production    bool   `env:"PRODUCTION" env-default:"true"`
	Version       string `env:"VERSION" env-default:"1"`
	MongoHost     string `env:"MONGO_HOST" env-default:"mongodb"`
	MongoPort     string `env:"MONGO_PORT" env-default:"27017"`
	MongoUsername string `env:"MONGO_USERNAME" env-required:"true"`
	MongoPassword string `env:"MONGO_PASSWORD" env-required:"true"`
}

var (
	config Config
	once   sync.Once
)

func Get() *Config {
	once.Do(func() {
		err := godotenv.Load()

		if err != nil {
			log.Println("error loading .env file")
		}
		err = cleanenv.ReadEnv(&config)
		if err != nil {
			panic(fmt.Sprintf("Failed to get config: %s", err))
		}
	})
	return &config
}

func (c *Config) GetMongoURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", c.MongoUsername, c.MongoPassword, c.MongoHost, c.MongoPort)
}
