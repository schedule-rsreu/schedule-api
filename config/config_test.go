package config_test

import (
	"testing"

	"github.com/schedule-rsreu/schedule-api/config"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	const postgresDSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	const dwhURL = "http://schedule-dwh.schedule-api.svc.cluster.local"
	t.Setenv("POSTGRES_DSN", postgresDSN)
	t.Setenv("DWH_URL", dwhURL)
	t.Setenv("TELEGRAM_BOT_TOKEN", "secret")
	t.Setenv("BANNED_IPS", "203.0.113.7,198.51.100.9")

	cfg := config.Get()
	t.Log(cfg)

	assert.Equal(t, "80", cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "1", cfg.Version)
	assert.True(t, cfg.Production)
	assert.Equal(t, postgresDSN, cfg.PostgresDSN)
	assert.Equal(t, dwhURL, cfg.DWHUrl)
	assert.Equal(t, "secret", cfg.TelegramBotToken)
	assert.Equal(t, []string{"203.0.113.7", "198.51.100.9"}, cfg.BannedIPs)
}
