package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/RitochitGhosh/olx-api/internal/config"
	"github.com/RitochitGhosh/olx-api/internal/db"
	"github.com/RitochitGhosh/olx-api/internal/handlers"
)

func main() {
	cfg := config.MustLoad()

	_, err := db.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect with database: %v", err)
	}

	fmt.Println("Connected to database...")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handlers.HealthHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	log.Printf("server is listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}

}
