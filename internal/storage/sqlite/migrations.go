package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	// Создаем таблицу migrations для отслеживания выполненных миграций
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var exists bool
		err = db.QueryRowContext(ctx, "SELECT 1 FROM migrations WHERE name = ?", name).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("checking migration %s: %w", name, err)
		}

		if err == nil {
			continue
		}

		b, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		stmts := strings.Split(string(b), ";")
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		for _, s := range stmts {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, s); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("migration %s failed: %w", name, err)
			}
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO migrations (name) VALUES (?)", name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
