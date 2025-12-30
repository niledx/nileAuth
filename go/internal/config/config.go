package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port       string
	DBAdapter  string
	SQLiteFile string
	JwtSecret  string
	LogLevel   string
	// PostgreSQL connection settings
	PostgresDSN      string
	PostgresHost    string
	PostgresPort    string
	PostgresUser    string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// BuildPostgresDSN constructs a PostgreSQL DSN from individual components or returns the provided DSN
func (c *Config) BuildPostgresDSN() (string, error) {
	// If DSN is provided directly, use it
	if c.PostgresDSN != "" {
		return c.PostgresDSN, nil
	}

	// Build DSN from individual components
	if c.PostgresHost == "" {
		return "", errors.New("POSTGRES_HOST or POSTGRES_DSN must be set")
	}
	if c.PostgresUser == "" {
		return "", errors.New("POSTGRES_USER must be set")
	}
	if c.PostgresDB == "" {
		return "", errors.New("POSTGRES_DB must be set")
	}

	port := c.PostgresPort
	if port == "" {
		port = "5432"
	}

	sslMode := c.PostgresSSLMode
	if sslMode == "" {
		sslMode = "disable" // Default to disable for local development
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		c.PostgresHost, port, c.PostgresUser, c.PostgresDB, sslMode)

	if c.PostgresPassword != "" {
		dsn += " password=" + c.PostgresPassword
	}

	return dsn, nil
}

func New() (*Config, error) {
	c := &Config{
		Port:       getenv("PORT", "8080"),
		DBAdapter:  getenv("DB_ADAPTER", "postgres"), // Default to postgres
		SQLiteFile: getenv("SQLITE_FILE", "./data/nile_go.db"),
		JwtSecret:  getenv("JWT_SECRET", "change-me"),
		LogLevel:   getenv("LOG_LEVEL", "info"),
		// PostgreSQL settings
		PostgresDSN:      getenv("POSTGRES_DSN", ""),
		PostgresHost:     getenv("POSTGRES_HOST", getenv("DB_HOST", "localhost")),
		PostgresPort:     getenv("POSTGRES_PORT", getenv("DB_PORT", "5432")),
		PostgresUser:     getenv("POSTGRES_USER", getenv("DB_USER", "nile")),
		PostgresPassword: getenv("POSTGRES_PASSWORD", getenv("DB_PASSWORD", "nilepass")),
		PostgresDB:       getenv("POSTGRES_DB", getenv("DB_NAME", "nileauth")),
		PostgresSSLMode:  getenv("POSTGRES_SSLMODE", getenv("DB_SSLMODE", "disable")),
	}

	// Validate PostgreSQL configuration if using postgres
	if c.DBAdapter == "postgres" {
		dsn, err := c.BuildPostgresDSN()
		if err != nil {
			return nil, fmt.Errorf("postgres configuration error: %w", err)
		}
		c.PostgresDSN = dsn
	}

	if c.DBAdapter == "sqlite" {
		// ensure sqlite file path is not empty
		if c.SQLiteFile == "" {
			return nil, errors.New("SQLITE_FILE must be set when DB_ADAPTER=sqlite")
		}
	}

	// Validate JWT secret in production
	env := strings.ToLower(getenv("NODE_ENV", getenv("ENV", "")))
	if env == "production" || env == "prod" {
		if c.JwtSecret == "" || c.JwtSecret == "change-me" {
			return nil, errors.New("JWT_SECRET must be set in production")
		}
	}

	// normalize port
	if _, err := strconv.Atoi(c.Port); err == nil {
		// ok
	} else {
		return nil, fmt.Errorf("invalid PORT: %s", c.Port)
	}

	return c, nil
}
