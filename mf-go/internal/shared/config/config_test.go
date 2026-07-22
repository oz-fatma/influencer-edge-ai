package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	for _, key := range []string{
		"PORT", "SERVER_PORT", "DB_PORT", "REDIS_PORT", "LOG_FORMAT", "DB_HOST", "REDIS_HOST",
	} {
		t.Setenv(key, "")
	}

	cfg := Load()

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "masterfabric", cfg.Database.User)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "db.example.com")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("DB_HOST")

	cfg := Load()
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
}

func TestLoad_PORTOverridesServerPort(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("SERVER_PORT", "8081")

	cfg := Load()
	assert.Equal(t, 10000, cfg.Server.Port)
}

func TestLoad_DBPoolInt32Bounds(t *testing.T) {
	os.Setenv("DB_MAX_CONNS", "50")
	os.Setenv("DB_MIN_CONNS", "2147483648")
	defer os.Unsetenv("DB_MAX_CONNS")
	defer os.Unsetenv("DB_MIN_CONNS")

	cfg := Load()
	assert.Equal(t, int32(50), cfg.Database.MaxConns)
	assert.Equal(t, int32(5), cfg.Database.MinConns)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}
	expected := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

func TestDatabaseConfig_DSN_WithSchema(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "testdb",
		SSLMode:  "require",
		Schema:   "mf",
	}
	dsn := cfg.DSN()
	assert.Contains(t, dsn, "sslmode=require")
	assert.Contains(t, dsn, "search_path")
	assert.Contains(t, dsn, "mf")
}

func TestDatabaseConfig_DSN_EscapesSpecialCharacters(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user@domain",
		Password: "p@ss:w?rd#",
		DBName:   "testdb",
		SSLMode:  "require",
	}
	dsn := cfg.DSN()
	assert.Contains(t, dsn, "postgres://")
	assert.Contains(t, dsn, "sslmode=require")
	assert.NotContains(t, dsn, "p@ss:w?rd#")
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := RedisConfig{Host: "redis.local", Port: 6380}
	assert.Equal(t, "redis.local:6380", cfg.Addr())
}

func TestLoad_DATABASE_URL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://edge_user:secret@dpg-example.frankfurt-postgres.render.com:5432/influencer_edge_db?sslmode=require")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")

	cfg := Load()

	assert.Equal(t, "dpg-example.frankfurt-postgres.render.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "edge_user", cfg.Database.User)
	assert.Equal(t, "secret", cfg.Database.Password)
	assert.Equal(t, "influencer_edge_db", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
}

func TestLoad_DATABASE_URL_DB_NAMEOverride(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://edge_user:secret@localhost:5432/other_db?sslmode=require")
	t.Setenv("DB_NAME", "influencer_edge_db")

	cfg := Load()
	assert.Equal(t, "influencer_edge_db", cfg.Database.DBName)
}

func TestLoad_REDIS_URL(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://red-abc123:6379")
	t.Setenv("REDIS_HOST", "")
	t.Setenv("REDIS_PORT", "")

	cfg := Load()
	assert.Equal(t, "redis://red-abc123:6379", cfg.Redis.URL)
}

func TestLoad_DB_SCHEMA(t *testing.T) {
	t.Setenv("DB_SCHEMA", "mf")

	cfg := Load()
	assert.Equal(t, "mf", cfg.Database.Schema)
}
