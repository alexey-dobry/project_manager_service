package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode,
	)
}

// JWTConfig — только то, что нужно для проверки. Никаких TTL.
type JWTConfig struct {
	Secret string
}

type LoggerConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.name", "groups-service")
	v.SetDefault("app.env", "dev")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", "8082")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.shutdown_timeout", "10s")

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", "5432")
	v.SetDefault("postgres.user", "postgres")
	v.SetDefault("postgres.password", "postgres")
	v.SetDefault("postgres.dbname", "groups")
	v.SetDefault("postgres.sslmode", "disable")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")

	cfg := &Config{
		App: AppConfig{
			Name: v.GetString("app.name"),
			Env:  v.GetString("app.env"),
		},
		HTTP: HTTPConfig{
			Host:            v.GetString("http.host"),
			Port:            v.GetString("http.port"),
			ReadTimeout:     v.GetDuration("http.read_timeout"),
			WriteTimeout:    v.GetDuration("http.write_timeout"),
			ShutdownTimeout: v.GetDuration("http.shutdown_timeout"),
		},
		Postgres: PostgresConfig{
			Host:     v.GetString("postgres.host"),
			Port:     v.GetString("postgres.port"),
			User:     v.GetString("postgres.user"),
			Password: v.GetString("postgres.password"),
			DBName:   v.GetString("postgres.dbname"),
			SSLMode:  v.GetString("postgres.sslmode"),
		},
		JWT: JWTConfig{Secret: v.GetString("jwt.secret")},
		Logger: LoggerConfig{
			Level:  v.GetString("logger.level"),
			Format: v.GetString("logger.format"),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required (must match auth-service)")
	}
	return cfg, nil
}
