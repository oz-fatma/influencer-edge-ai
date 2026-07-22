package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Kafka     KafkaConfig
	WebSocket WebSocketConfig
	Log       LogConfig
}

// WebSocketConfig holds real-time WebSocket settings.
type WebSocketConfig struct {
	Enabled         bool
	MaxConnections  int
	PingIntervalSec int
	ReadBufferSize  int
	WriteBufferSize int
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	CORSAllowedOrigins []string
	MaxBodyBytes      int64
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	// Schema is an optional PostgreSQL schema (e.g. "mf" on shared influencer_edge_db).
	// When set, connections use search_path schema,public so legacy public.* tables stay untouched.
	Schema   string
	MaxConns int32
	MinConns int32
}

// DSN returns the PostgreSQL connection string with escaped credentials.
func (d DatabaseConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   "/" + d.DBName,
	}
	u.RawQuery = url.Values{"sslmode": {d.SSLMode}}.Encode()
	return u.String()
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	URL      string // Render/Heroku REDIS_URL (preferred when set)
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

// KafkaConfig holds Kafka connection and consumer settings.
type KafkaConfig struct {
	Brokers           []string
	GroupID           string
	Enabled           bool
	NumPartitions     int
	ReplicationFactor int
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:               envOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:               serverPort(),
			ReadTimeout:        time.Duration(envOrDefaultInt("SERVER_READ_TIMEOUT_SECONDS", 15)) * time.Second,
			WriteTimeout:       time.Duration(envOrDefaultInt("SERVER_WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
			IdleTimeout:        time.Duration(envOrDefaultInt("SERVER_IDLE_TIMEOUT_SECONDS", 60)) * time.Second,
			CORSAllowedOrigins: envOrDefaultSlice("CORS_ALLOWED_ORIGINS", nil),
			MaxBodyBytes:       envOrDefaultInt64("MAX_BODY_BYTES", 1<<20),
		},
		Database: loadDatabaseConfig(),
		Redis: loadRedisConfig(),
		JWT: JWTConfig{
			Secret:          envOrDefault("JWT_SECRET", "change-me-in-production"),
			ExpirationHours: envOrDefaultInt("JWT_EXPIRATION_HOURS", 24),
			Issuer:          envOrDefault("JWT_ISSUER", "masterfabric"),
		},
		Kafka: KafkaConfig{
			Brokers:           envOrDefaultSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			GroupID:           envOrDefault("KAFKA_GROUP_ID", "masterfabric-go"),
			Enabled:           envOrDefault("KAFKA_ENABLED", "false") == "true",
			NumPartitions:     envOrDefaultInt("KAFKA_NUM_PARTITIONS", 3),
			ReplicationFactor: envOrDefaultInt("KAFKA_REPLICATION_FACTOR", 1),
		},
		WebSocket: WebSocketConfig{
			Enabled:         envOrDefault("WS_ENABLED", "true") == "true",
			MaxConnections:  envOrDefaultInt("WS_MAX_CONNECTIONS", 1000),
			PingIntervalSec: envOrDefaultInt("WS_PING_INTERVAL_SECONDS", 30),
			ReadBufferSize:  envOrDefaultInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize: envOrDefaultInt("WS_WRITE_BUFFER_SIZE", 1024),
		},
		Log: LogConfig{
			Level:  envOrDefault("LOG_LEVEL", "info"),
			Format: envOrDefault("LOG_FORMAT", "json"),
		},
	}
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func envOrDefaultInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func envOrDefaultInt32(key string, defaultVal int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(n)
		}
	}
	return defaultVal
}

func envOrDefaultInt64(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}

func envOrDefaultSlice(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		parts := strings.Split(val, ",")
		var result []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultVal
}

// loadDatabaseConfig reads DB settings from DATABASE_URL (Render/Heroku) or DB_* vars.
// Individual DB_* vars override parsed DATABASE_URL values when set.
func loadDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{
		Host:     envOrDefault("DB_HOST", "localhost"),
		Port:     envOrDefaultInt("DB_PORT", 5432),
		User:     envOrDefault("DB_USER", "masterfabric"),
		Password: envOrDefault("DB_PASSWORD", "masterfabric"),
		DBName:   envOrDefault("DB_NAME", "masterfabric"),
		SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
		MaxConns: envOrDefaultInt32("DB_MAX_CONNS", 25),
		MinConns: envOrDefaultInt32("DB_MIN_CONNS", 5),
	}

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		if parsed, err := parseDatabaseURL(dsn); err == nil {
			cfg = parsed
		}
	}

	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.SSLMode = v
	}
	if v := os.Getenv("DB_SCHEMA"); v != "" {
		cfg.Schema = v
	}
	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			cfg.MaxConns = int32(n)
		}
	}
	if v := os.Getenv("DB_MIN_CONNS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			cfg.MinConns = int32(n)
		}
	}

	return cfg
}

func parseDatabaseURL(raw string) (DatabaseConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return DatabaseConfig{}, err
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return DatabaseConfig{}, fmt.Errorf("unsupported database scheme: %s", u.Scheme)
	}

	cfg := DatabaseConfig{
		SSLMode:  "require",
		MaxConns: 25,
		MinConns: 5,
	}

	if u.User != nil {
		cfg.User = u.User.Username()
		if password, ok := u.User.Password(); ok {
			cfg.Password = password
		}
	}

	if u.Hostname() != "" {
		cfg.Host = u.Hostname()
	}
	if port := u.Port(); port != "" {
		if parsedPort, err := strconv.Atoi(port); err == nil {
			cfg.Port = parsedPort
		}
	} else {
		cfg.Port = 5432
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName != "" {
		cfg.DBName = dbName
	}

	if sslMode := u.Query().Get("sslmode"); sslMode != "" {
		cfg.SSLMode = sslMode
	}

	return cfg, nil
}

// loadRedisConfig reads Redis settings from REDIS_URL (Render Key Value) or REDIS_* vars.
func loadRedisConfig() RedisConfig {
	cfg := RedisConfig{
		Host:     envOrDefault("REDIS_HOST", "localhost"),
		Port:     envOrDefaultInt("REDIS_PORT", 6379),
		Password: envOrDefault("REDIS_PASSWORD", ""),
		DB:       envOrDefaultInt("REDIS_DB", 0),
	}

	if raw := os.Getenv("REDIS_URL"); raw != "" {
		cfg.URL = raw
	}

	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			cfg.DB = db
		}
	}

	return cfg
}

// serverPort prefers Render/Heroku-style PORT, then SERVER_PORT, then 8080.
func serverPort() int {
	if port := os.Getenv("PORT"); port != "" {
		if intVal, err := strconv.Atoi(port); err == nil {
			return intVal
		}
	}
	return envOrDefaultInt("SERVER_PORT", 8080)
}
