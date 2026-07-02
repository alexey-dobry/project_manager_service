// Package main — точка входа groups-service.
//
// @title           Groups Service API
// @version         1.0
// @description     Управление студенческими группами и членством в них.
// @BasePath        /
// @securityDefinitions.apikey  BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/student-pm/groups-service/docs" // swagger docs

	"github.com/student-pm/groups-service/internal/config"
	httpdelivery "github.com/student-pm/groups-service/internal/delivery/http"
	"github.com/student-pm/groups-service/internal/pkg/authclient"
	jwtpkg "github.com/student-pm/groups-service/internal/pkg/jwt"
	"github.com/student-pm/groups-service/internal/pkg/logger"
	"github.com/student-pm/groups-service/internal/pkg/validator"
	"github.com/student-pm/groups-service/internal/repository"
	"github.com/student-pm/groups-service/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Logger.Level, cfg.Logger.Format)
	log.Info().Str("env", cfg.App.Env).Msg("starting groups-service")

	if err := runMigrations(cfg.Postgres.DSN(), "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Info().Msg("migrations applied")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("pgx pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("pg ping: %w", err)
	}

	repo := repository.NewPostgresRepo(pool)
	verifier, err := jwtpkg.New(cfg.JWT.Secret)
	if err != nil {
		return fmt.Errorf("jwt verifier: %w", err)
	}
	svc := usecase.NewGroupService(repo, time.Now)
	v := validator.New()
	authClient := authclient.New(cfg.AuthServiceURL)
	h := httpdelivery.NewHandler(svc, v, authClient)

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		BodyLimit:    1 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := http.StatusInternalServerError
			msg := "internal server error"
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
				msg = fe.Message
			}
			if code >= http.StatusInternalServerError {
				log.Error().Err(err).Str("path", c.Path()).Msg("unhandled error")
			} else {
				log.Warn().Err(err).Str("path", c.Path()).Int("status", code).Msg("request error")
			}
			return c.Status(code).JSON(fiber.Map{
				"error": fiber.Map{"code": httpErrCode(code), "message": msg},
			})
		},
	})
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(httpdelivery.RequestID())
	app.Use(httpdelivery.AccessLog(log))

	httpdelivery.RegisterRoutes(app, h, verifier)

	addr := cfg.HTTP.Host + ":" + cfg.HTTP.Port
	srvErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", addr).Msg("http server listening")
		if err := app.Listen(addr); err != nil {
			srvErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		return err
	case sig := <-stop:
		log.Info().Str("signal", sig.String()).Msg("shutting down")
	}

	shCtx, shCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shCancel()
	if err := app.ShutdownWithContext(shCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Info().Msg("bye")
	return nil
}

func runMigrations(dsn, dir string) error {
	m, err := migrate.New("file://"+dir, dsn)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// httpErrCode переводит HTTP-статус в короткий машиночитаемый код ошибки.
// Используется как fallback, когда конкретный доменный код неизвестен —
// например, для ошибок самого фреймворка (404 на несуществующий маршрут,
// 405 на неподдерживаемый метод, 413 при превышении BodyLimit).
func httpErrCode(status int) string {
	switch status {
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusRequestEntityTooLarge:
		return "payload_too_large"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusTooManyRequests:
		return "rate_limited"
	default:
		if status >= 500 {
			return "internal_error"
		}
		return "error"
	}
}
