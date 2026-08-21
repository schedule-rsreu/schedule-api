package config_test

import (
	"testing"

	"github.com/schedule-rsreu/schedule-api/config"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	const postgresDSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	t.Setenv("POSTGRES_DSN", postgresDSN)

	cfg := config.Get()
	t.Log(cfg)

	assert.Equal(t, "80", cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "1", cfg.Version)
	assert.True(t, cfg.Production)
	assert.Equal(t, postgresDSN, cfg.PostgresDSN)
}
