package config

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Port          string `env:"PORT"           env-default:"80"`
	Host          string `env:"HOST"           env-default:"0.0.0.0"`
	Version       string `env:"VERSION"        env-default:"1"`
	MongoHost     string `env:"MONGO_HOST"     env-default:"mongodb"`
	MongoPort     string `env:"MONGO_PORT"     env-default:"27017"`
	MongoUsername string `env:"MONGO_USERNAME"                       env-required:"true"`
	MongoPassword string `env:"MONGO_PASSWORD"                       env-required:"true"`
	MongoDBName   string `env:"MONGO_DB_NAME"                        env-required:"true"`
	PostgresDSN   string `env:"POSTGRES_DSN"                         env-required:"true"`
	Production    bool   `env:"PRODUCTION"     env-default:"true"`
}

var (
	config Config    //nolint:gochecknoglobals,lll // Global config is initialized once and accessed throughout the application.
	once   sync.Once //nolint:gochecknoglobals,lll // Ensures the config is initialized only once, which requires a global sync.Once.
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
	hostPort := net.JoinHostPort(c.MongoHost, c.MongoPort)
	return fmt.Sprintf("mongodb://%s:%s@%s", c.MongoUsername, c.MongoPassword, hostPort)
}
