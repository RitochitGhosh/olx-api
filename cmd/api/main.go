package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/RitochitGhosh/olx-api/internal/config"
	"github.com/RitochitGhosh/olx-api/internal/db"
	"github.com/RitochitGhosh/olx-api/internal/handlers"
	"github.com/RitochitGhosh/olx-api/internal/middleware"
)

func main() {
	cfg := config.MustLoad()

	db, err := db.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect with database: %v", err)
	}
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		// ReplaceAttr: strips sensitive key-value, eg. password, auth-key,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	fmt.Println("Connected to database...")

	lh := handlers.NewListingHandler(db, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handlers.HealthHandler)
	mux.HandleFunc("GET /listings", lh.FetchListings)
	mux.HandleFunc("POST /listings", lh.CreateListing)
	mux.HandleFunc("DELETE /listings/{id}", lh.DeleteListing)

	handler := middleware.RequestId(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	log.Printf("server is listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
