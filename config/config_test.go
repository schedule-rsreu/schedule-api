package config_test

import (
	"testing"

	"github.com/schedule-rsreu/schedule-api/config"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	assert.Panics(t, func() {
		config.Get()
	}, "config.Get() should panic")

	t.Setenv("MONGO_USERNAME", "mongo")

	t.Setenv("MONGO_PASSWORD", "mongo")
	t.Setenv("MONGO_DB_NAME", "mongo")

	cfg := config.Get()
	t.Log(cfg)

	assert.Equal(t, "80", cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "1", cfg.Version)
	assert.Equal(t, "mongodb", cfg.MongoHost)
	assert.Equal(t, "27017", cfg.MongoPort)
	assert.True(t, cfg.Production)
	assert.Equal(t, "mongodb://mongo:mongo@mongodb:27017", cfg.GetMongoURI())
	assert.Equal(t, "mongo", cfg.MongoDBName)
}
