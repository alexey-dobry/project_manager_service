package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/student-pm/api-gateway/internal/config"
	httpdelivery "github.com/student-pm/api-gateway/internal/delivery/http"
	jwtpkg "github.com/student-pm/api-gateway/internal/pkg/jwt"
	"github.com/student-pm/api-gateway/internal/pkg/logger"
	"github.com/student-pm/api-gateway/internal/proxy"
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
	log.Info().Str("env", cfg.App.Env).Msg("starting api-gateway")

	verifier, err := jwtpkg.New(cfg.JWT.Secret)
	if err != nil {
		return fmt.Errorf("jwt verifier: %w", err)
	}

	// Три proxy.
	authProxy, err := proxy.New(cfg.Upstreams.Auth, cfg.Upstreams.Timeout, log)
	if err != nil {
		return fmt.Errorf("auth proxy: %w", err)
	}
	groupsProxy, err := proxy.New(cfg.Upstreams.Groups, cfg.Upstreams.Timeout, log)
	if err != nil {
		return fmt.Errorf("groups proxy: %w", err)
	}
	projectsProxy, err := proxy.New(cfg.Upstreams.Projects, cfg.Upstreams.Timeout, log)
	if err != nil {
		return fmt.Errorf("projects proxy: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		ReadTimeout:           cfg.HTTP.ReadTimeout,
		WriteTimeout:          cfg.HTTP.WriteTimeout,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error().Err(err).Str("path", c.Path()).Msg("unhandled error")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{"code": "internal_error", "message": "internal server error"},
			})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowCredentials: false, // меняйте на true, если нужны cookies
	}))
	app.Use(httpdelivery.RequestID())
	app.Use(httpdelivery.AccessLog(log))
	// Rate limiting — N запросов с одного IP за окно.
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimit.Max,
		Expiration: cfg.RateLimit.Duration,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"error": fiber.Map{"code": "rate_limited", "message": "too many requests"},
			})
		},
	}))

	httpdelivery.RegisterRoutes(app, httpdelivery.Upstreams{
		Auth:     authProxy,
		Groups:   groupsProxy,
		Projects: projectsProxy,
	}, verifier, httpdelivery.SwaggerHandler())

	addr := cfg.HTTP.Host + ":" + cfg.HTTP.Port
	srvErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", addr).Msg("gateway listening")
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
