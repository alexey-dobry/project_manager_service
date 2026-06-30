package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Logger    LoggerConfig
	Upstreams Upstreams
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

type JWTConfig struct {
	Secret string
}

type RateLimitConfig struct {
	Max      int
	Duration time.Duration
}

type CORSConfig struct {
	AllowOrigins string // "*" или CSV
}

type LoggerConfig struct {
	Level  string
	Format string
}

// Upstreams — адреса трёх backend-сервисов.
type Upstreams struct {
	Auth     string
	Groups   string
	Projects string
	// Таймаут одного proxy-запроса.
	Timeout time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.name", "api-gateway")
	v.SetDefault("app.env", "dev")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", "8080")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.shutdown_timeout", "10s")

	v.SetDefault("rate_limit.max", 100)
	v.SetDefault("rate_limit.duration", "1m")

	v.SetDefault("cors.allow_origins", "*")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")

	// upstream-дефолты — имена сервисов в docker-compose
	v.SetDefault("upstream.auth", "http://auth-service:8081")
	v.SetDefault("upstream.groups", "http://groups-service:8082")
	v.SetDefault("upstream.projects", "http://projects-service:8083")
	v.SetDefault("upstream.timeout", "15s")

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
		JWT:       JWTConfig{Secret: v.GetString("jwt.secret")},
		RateLimit: RateLimitConfig{Max: v.GetInt("rate_limit.max"), Duration: v.GetDuration("rate_limit.duration")},
		CORS:      CORSConfig{AllowOrigins: v.GetString("cors.allow_origins")},
		Logger:    LoggerConfig{Level: v.GetString("logger.level"), Format: v.GetString("logger.format")},
		Upstreams: Upstreams{
			Auth:     v.GetString("upstream.auth"),
			Groups:   v.GetString("upstream.groups"),
			Projects: v.GetString("upstream.projects"),
			Timeout:  v.GetDuration("upstream.timeout"),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	for k, v := range map[string]string{
		"upstream.auth": cfg.Upstreams.Auth,
		"upstream.groups": cfg.Upstreams.Groups,
		"upstream.projects": cfg.Upstreams.Projects,
	} {
		if v == "" {
			return nil, fmt.Errorf("%s is required", k)
		}
	}
	return cfg, nil
}
