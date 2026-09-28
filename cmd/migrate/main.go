package main

import (
	"fmt"
	"log"
	"os"

	"github.com/RitochitGhosh/olx-api/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: make migrate-<up | dowm>")
	}

	cfg := config.MustLoad()

	m, err := migrate.New(
		"file://migrations",
		cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("migtation.new: %v", err)
	}

	fmt.Println("running migrations...")
	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil {
			log.Fatalf("migration.up: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("migration.up: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
	fmt.Println("done running migration...")
}
