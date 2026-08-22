package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string `env:"PORT"         env-default:"80"`
	Host        string `env:"HOST"         env-default:"0.0.0.0"`
	Version     string `env:"VERSION"      env-default:"1"`
	PostgresDSN string `env:"POSTGRES_DSN"                          env-required:"true"`
	Environment string `env:"ENVIRONMENT"  env-default:"prod"`
	OtlEndpoint string `env:"OTL_ENDPOINT" env-default:"tempo:4317"`
	DWHUrl      string `env:"DWH_URL"                               env-required:"true"`
	Production  bool   `env:"PRODUCTION"   env-default:"true"`
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
