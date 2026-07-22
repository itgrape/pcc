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
	Casdoor   CasdoorConfig
	Snowflake SnowflakeConfig
}

type ServerConfig struct {
	Addr string
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type CasdoorConfig struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
	Organization string
	Application  string
}

type SnowflakeConfig struct {
	NodeID int64
}

func Load() (*Config, error) {
	_ = loadDotEnv()

	cfg := &Config{
		Server: ServerConfig{
			Addr: env("PCC_SERVER_ADDR", ":8080"),
		},
		DB: DBConfig{
			DSN:             env("PCC_DB_DSN", ""),
			MaxOpenConns:    envInt("PCC_DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:    envInt("PCC_DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: envDuration("PCC_DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Casdoor: CasdoorConfig{
			Endpoint:     env("PCC_CASDOOR_ENDPOINT", ""),
			ClientID:     env("PCC_CASDOOR_CLIENT_ID", ""),
			ClientSecret: env("PCC_CASDOOR_CLIENT_SECRET", ""),
			Organization: env("PCC_CASDOOR_ORGANIZATION", ""),
			Application:  env("PCC_CASDOOR_APPLICATION", ""),
		},
		Snowflake: SnowflakeConfig{
			NodeID: int64(envInt("PCC_SNOWFLAKE_NODE_ID", 1)),
		},
	}

	if cfg.DB.DSN == "" {
		return nil, fmt.Errorf("config: PCC_DB_DSN is required")
	}
	if cfg.Casdoor.Endpoint == "" || cfg.Casdoor.ClientID == "" ||
		cfg.Casdoor.ClientSecret == "" || cfg.Casdoor.Organization == "" ||
		cfg.Casdoor.Application == "" {
		return nil, fmt.Errorf("config: PCC_CASDOOR_ENDPOINT, PCC_CASDOOR_CLIENT_ID, PCC_CASDOOR_CLIENT_SECRET, PCC_CASDOOR_ORGANIZATION and PCC_CASDOOR_APPLICATION are required")
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
