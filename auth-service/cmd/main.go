// Package main — точка входа auth-service.
//
// @title           Auth Service API
// @version         1.0
// @description     Сервис аутентификации и пользователей системы управления студенческими проектами.
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

	_ "github.com/student-pm/auth-service/docs" // swagger docs

	"github.com/student-pm/auth-service/internal/config"
	httpdelivery "github.com/student-pm/auth-service/internal/delivery/http"
	"github.com/student-pm/auth-service/internal/pkg/hasher"
	jwtpkg "github.com/student-pm/auth-service/internal/pkg/jwt"
	"github.com/student-pm/auth-service/internal/pkg/logger"
	"github.com/student-pm/auth-service/internal/pkg/validator"
	"github.com/student-pm/auth-service/internal/repository"
	"github.com/student-pm/auth-service/internal/usecase"
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
	log.Info().Str("env", cfg.App.Env).Msg("starting auth-service")

	// 1. Применяем миграции до старта пула.
	if err := runMigrations(cfg.Postgres.DSN(), "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Info().Msg("migrations applied")

	// 2. Пул подключений.
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

	// 3. DI: репозитории → usecase → handlers.
	repo := repository.NewPostgresRepo(pool)
	hash := hasher.New(cfg.Bcrypt.Cost)
	tp, err := jwtpkg.New(jwtpkg.Config{
		Secret:     cfg.JWT.Secret,
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
		Issuer:     cfg.JWT.Issuer,
	})
	if err != nil {
		return fmt.Errorf("jwt provider: %w", err)
	}
	auth := usecase.NewAuthService(repo, repo, hash, tp, time.Now)
	v := validator.New()
	h := httpdelivery.NewHandler(auth, v)

	// 4. Fiber.
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Fallback на случай неперехваченной ошибки.
			log.Error().Err(err).Str("path", c.Path()).Msg("unhandled error")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{"code": "internal_error", "message": "internal server error"},
			})
		},
	})
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(httpdelivery.RequestID())
	app.Use(httpdelivery.AccessLog(log))

	httpdelivery.RegisterRoutes(app, h, tp)

	// 5. Graceful shutdown.
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

// runMigrations применяет up-миграции из папки.
// Если изменений нет (ErrNoChange) — это норма, возвращаем nil.
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
