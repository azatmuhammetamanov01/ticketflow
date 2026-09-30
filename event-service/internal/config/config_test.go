package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withTempWorkingDir(t *testing.T) string {
	t.Helper()

	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	return tempDir
}

func TestLoad_UsesDefaultValuesWhenEnvMissing(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("SERVER_HOST", "")
	t.Setenv("HTTP_HOST", "")
	t.Setenv("APP_ENV", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "postgres", cfg.Database.Password)
	assert.Equal(t, "product_db", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "8080", cfg.Server.HTTP_Port)
	assert.Equal(t, "9091", cfg.Server.GRPC_Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "development", cfg.App.Environment)
}

func TestLoad_UsesEnvironmentValues(t *testing.T) {
	withTempWorkingDir(t)

	t.Setenv("DB_HOST", "db.example.local")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "tester")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "events_test")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("HTTP_PORT", "8088")
	t.Setenv("GRPC_PORT", "9098")
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("APP_ENV", "test")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "db.example.local", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "tester", cfg.Database.User)
	assert.Equal(t, "secret", cfg.Database.Password)
	assert.Equal(t, "events_test", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "8088", cfg.Server.HTTP_Port)
	assert.Equal(t, "9098", cfg.Server.GRPC_Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "test", cfg.App.Environment)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "1234",
		DBName:   "event_db",
		SSLMode:  "disable",
	}

	assert.Equal(t, "host=localhost port=5432 user=postgres password=1234 dbname=event_db sslmode=disable", cfg.DSN())
}
