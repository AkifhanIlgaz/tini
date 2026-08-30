// Command server runs the application's HTTP server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AkifhanIlgaz/tini/internal/config"
	"github.com/AkifhanIlgaz/tini/internal/features/auth"
	"github.com/AkifhanIlgaz/tini/internal/features/dashboard"
	"github.com/AkifhanIlgaz/tini/internal/features/user"
	"github.com/AkifhanIlgaz/tini/internal/features/venue"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/logging"
	db "github.com/AkifhanIlgaz/tini/internal/platform/mongo"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	slogfiber "github.com/samber/slog-fiber"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// No logger yet — config itself failed to load, so fall back to
		// slog's unconfigured default (plain text to stderr).
		slog.Error("config.Load", "error", err)
		os.Exit(1)
	}

	logger := logging.New(cfg)
	slog.SetDefault(logger)

	ctx := context.Background()

	mongoClient, err := db.Connect(ctx, cfg.Mongo)
	if err != nil {
		slog.Error("mongo.Connect", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			slog.Error("mongo.Disconnect", "error", err)
		}
	}()

	usersRepo, err := user.NewRepository(mongoClient)
	if err != nil {
		slog.Error("user.NewRepository", "error", err)
		os.Exit(1)
	}

	authHandler := auth.NewAuthHandler(usersRepo)
	authHandler.RegisterProviders(cfg)

	sessionStore := session.NewStore(cfg)

	app := fiber.New()
	app.Use(slogfiber.New(logger))
	app.Use(session.New(sessionStore))
	app.Use(csrf.New(cfg, sessionStore))
	app.Use("/static", static.New("./static"))

	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	userService := user.NewService(usersRepo)

	authHandler.RegisterRoutes(app)
	dashboard.NewHandler().RegisterRoutes(app)
	user.NewHandler(userService).RegisterRoutes(app)
	venue.NewHandler().RegisterRoutes(app)

	go func() {
		slog.Info("server listening", "port", cfg.Server.Port, "env", cfg.Env)
		// DisableStartupMessage: Fiber's ASCII banner is a plain
		// fmt.Println, not a structured log — it would sit oddly next to
		// JSON output in production. The Info line above already covers
		// the same info.
		if err := app.Listen(":"+cfg.Server.Port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
			slog.Error("app.Listen", "error", err)
			os.Exit(1)
		}
	}()

	// air (and any other process manager) sends SIGINT/SIGTERM on rebuild
	// and expects the port back — without this, app.Listen never returns,
	// the process lingers past its parent, and the next build fails with
	// "address already in use".
	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-shutdownCtx.Done()
	stop()

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		slog.Error("app.Shutdown", "error", err)
	}
}
