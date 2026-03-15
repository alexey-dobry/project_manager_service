package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config — все настройки сервиса.
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Postgres PostgresConfig
	JWT      JWTConfig
	Bcrypt   BcryptConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Name string
	Env  string // dev | prod
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

// DSN формирует строку подключения для pgx.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode,
	)
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
}

type BcryptConfig struct {
	Cost int
}

type LoggerConfig struct {
	Level  string
	Format string // json | console
}

// Load читает .env (если есть) и переменные окружения.
func Load() (*Config, error) {
	_ = godotenv.Load() // отсутствие .env — не ошибка (в проде их и не должно быть)

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// дефолты
	v.SetDefault("app.name", "auth-service")
	v.SetDefault("app.env", "dev")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", "8081")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.shutdown_timeout", "10s")

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", "5432")
	v.SetDefault("postgres.user", "postgres")
	v.SetDefault("postgres.password", "postgres")
	v.SetDefault("postgres.dbname", "auth")
	v.SetDefault("postgres.sslmode", "disable")

	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "720h") // 30 дней
	v.SetDefault("jwt.issuer", "auth-service")

	v.SetDefault("bcrypt.cost", 10)

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
		JWT: JWTConfig{
			Secret:     v.GetString("jwt.secret"),
			AccessTTL:  v.GetDuration("jwt.access_ttl"),
			RefreshTTL: v.GetDuration("jwt.refresh_ttl"),
			Issuer:     v.GetString("jwt.issuer"),
		},
		Bcrypt: BcryptConfig{Cost: v.GetInt("bcrypt.cost")},
		Logger: LoggerConfig{
			Level:  v.GetString("logger.level"),
			Format: v.GetString("logger.format"),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}
