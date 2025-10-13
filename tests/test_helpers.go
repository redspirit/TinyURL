package tests

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"os"
	"testing"
	"tinyurl/internal/repository"
	"tinyurl/internal/service/link"
	"tinyurl/internal/storage/sqlite"
	sqliteRepo "tinyurl/internal/storage/sqlite/repository"
)

type TestDeps struct {
	Svc *link.Service
	Rep repository.LinkRepository
	Log *slog.Logger
	DB  *sql.DB
}

func NewTestDeps(t *testing.T) *TestDeps {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	repo := sqliteRepo.NewLinkRepo(db)
	service := link.New(repo)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	t.Cleanup(func() {
		db.Close()
		os.Remove("test.db")
		os.Remove("test.db-shm")
		os.Remove("test.db-wal")
	})

	return &TestDeps{
		Svc: service,
		Rep: repo,
		Log: logger,
		DB:  db,
	}
}

func setupTestDB() (*sql.DB, error) {
	os.Remove("test.db")
	os.Remove("test.db-shm")
	os.Remove("test.db-wal")

	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func setupTestService() (*link.Service, func(), error) {
	db, err := setupTestDB()
	if err != nil {
		return nil, nil, err
	}

	repo := sqliteRepo.NewLinkRepo(db)
	service := link.New(repo)

	cleanup := func() {
		db.Close()
		os.Remove("test.db")
		os.Remove("test.db-shm")
		os.Remove("test.db-wal")
	}

	return service, cleanup, nil
}
