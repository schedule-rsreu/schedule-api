package config_test

import (
	"testing"

	"github.com/schedule-rsreu/schedule-api/config"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigReturnsDefaultValues(t *testing.T) {
	// Remove any existing .env file
	// os.Remove(".env")

	cfg := config.Get()

	assert.Equal(t, "80", cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "1", cfg.Version)
	assert.Equal(t, "mongodb", cfg.MongoHost)
	assert.Equal(t, "27017", cfg.MongoPort)
	assert.True(t, cfg.Production)
	assert.Equal(t, "mongodb://mongo:mongo@mongodb:27017", cfg.GetMongoURI())
}
