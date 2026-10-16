package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	AdminAuth AdminAuthConfig
	Snowflake SnowflakeConfig
}

type ServerConfig struct {
	AdminAddr   string
	RuntimeAddr string
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AdminAuthConfig struct {
	Username     string
	Password     string
	SecureCookie bool
}

type SnowflakeConfig struct {
	NodeID int64
}

func Load() (*Config, error) {
	_ = loadDotEnv()

	cfg := &Config{
		Server: ServerConfig{
			AdminAddr:   env("PCC_ADMIN_ADDR", ":8080"),
			RuntimeAddr: env("PCC_RUNTIME_ADDR", ":8081"),
		},
		DB: DBConfig{
			DSN:             env("PCC_DB_DSN", ""),
			MaxOpenConns:    envInt("PCC_DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:    envInt("PCC_DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: envDuration("PCC_DB_CONN_MAX_LIFETIME", time.Hour),
		},
		AdminAuth: AdminAuthConfig{
			Username:     env("PCC_ADMIN_USERNAME", ""),
			Password:     env("PCC_ADMIN_PASSWORD", ""),
			SecureCookie: env("PCC_COOKIE_SECURE", "true") != "false",
		},
		Snowflake: SnowflakeConfig{
			NodeID: int64(envInt("PCC_SNOWFLAKE_NODE_ID", 1)),
		},
	}

	if cfg.DB.DSN == "" {
		return nil, fmt.Errorf("config: PCC_DB_DSN is required")
	}
	if cfg.AdminAuth.Username == "" || len(cfg.AdminAuth.Password) < 16 {
		return nil, fmt.Errorf("config: PCC_ADMIN_USERNAME and PCC_ADMIN_PASSWORD (at least 16 characters) are required")
	}
	return cfg, nil
}

func loadDotEnv() error {
	path := os.Getenv("PCC_ENV_FILE")
	if path == "" {
		path = ".env"
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	return godotenv.Load(path)
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
