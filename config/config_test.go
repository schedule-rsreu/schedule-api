package config_test

import (
	"os"
	"testing"

	"github.com/schedule-rsreu/schedule-api/config"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigReturnsDefaultValues(t *testing.T) {
	if err := os.Setenv("MONGO_USERNAME", "mongo"); err != nil {
		return
	}

	if err := os.Setenv("MONGO_PASSWORD", "mongo"); err != nil {
		return
	}

	cfg := config.Get()
	t.Log(cfg)

	assert.Equal(t, "80", cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "1", cfg.Version)
	assert.Equal(t, "mongodb", cfg.MongoHost)
	assert.Equal(t, "27017", cfg.MongoPort)
	assert.True(t, cfg.Production)
	assert.Equal(t, "mongodb://mongo:mongo@mongodb:27017", cfg.GetMongoURI())

	if err := os.Unsetenv("MONGO_USERNAME"); err != nil {
		return
	}

	if err := os.Unsetenv("MONGO_PASSWORD"); err != nil {
		return
	}
}
