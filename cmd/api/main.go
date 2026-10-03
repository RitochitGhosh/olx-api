package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RitochitGhosh/olx-api/internal/config"
	"github.com/RitochitGhosh/olx-api/internal/db"
	"github.com/RitochitGhosh/olx-api/internal/handlers"
	"github.com/RitochitGhosh/olx-api/internal/listing"
	"github.com/RitochitGhosh/olx-api/internal/middleware"
)

func main() {
	// Config
	cfg := config.MustLoad()

	// Logger
	logHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		},
	)

	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	// Database
	database, err := db.Connect(cfg.DatabaseUrl)
	if err != nil {
		logger.Error(
			"failed to connect to database",
			"error", err,
		)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("connected to database")

	// Dependencies
	listingRepo := listing.NewPostgresRepository(database)
	listingService := listing.NewService(
		listingRepo,
	)
	listingHandler := listing.NewHandler(
		listingService,
		logger,
	)

	// Router

	mux := http.NewServeMux()

	listing.RegisterRoutes(
		mux,
		listingHandler,
	)

	mux.HandleFunc(
		"GET /healthz",
		handlers.HealthHandler,
	)

	// Middleware
	var handler http.Handler = mux
	handler = middleware.RequestId(handler)

	// Later:
	// handler = middleware.RequestLogger(logger)(handler)
	// handler = middleware.Recover(logger)(handler)

	// HTTP Server

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info(
			"server started",
			"address", server.Addr,
		)

		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"server failed",
				"error", err,
			)

			os.Exit(1)
		}
	}()

	// Graceful shutdown
	shutdownSignal := make(chan os.Signal, 1)

	signal.Notify(
		shutdownSignal,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-shutdownSignal

	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(
			"server shutdown failed",
			"error", err,
		)
		return
	}

	logger.Info("server stopped gracefully")
}
