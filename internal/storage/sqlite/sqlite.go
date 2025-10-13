package sqlite

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type DB struct {
	*sql.DB
}

func Open(dsn string) (*DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &DB{DB: db}, nil
}
