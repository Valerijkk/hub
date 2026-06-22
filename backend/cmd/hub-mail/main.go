package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	fiberrecover "github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/valerii/hub/backend/internal/mail"
)

func main() {
	cfg := mail.LoadConfig()
	svc := mail.NewService(cfg)
	handler := mail.NewHubHandler(svc)

	app := fiber.New(fiber.Config{
		AppName: "hub-mail",
	})
	app.Use(fiberrecover.New())

	handler.RegisterRoutes(app)

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("hub-mail starting", "addr", addr, "resend_ready", cfg.ResendAPIKey != "")

	go func() {
		if err := app.Listen(addr); err != nil {
			slog.Error("hub-mail listen failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("hub-mail shutting down")
	_ = app.Shutdown()
}
