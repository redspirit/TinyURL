package main

import (
	"context"
	"flag"
	"log"

	"tinyurl/internal/config"
	"tinyurl/internal/storage/sqlite"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	dsn := flag.String("dsn", cfg.DB.DSN, "SQLite DSN (file path)")
	flag.Parse()

	db, err := sqlite.Open(*dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := sqlite.Migrate(context.Background(), db.DB); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations applied successfully")
}
