package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"tinyurl/internal/repository"
)

type LinkRepo struct {
	db *sql.DB
}

func NewLinkRepo(db *sql.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) Create(ctx context.Context, l *repository.Link) error {
	var expires interface{}
	if l.ExpiresAt != nil {
		expires = *l.ExpiresAt
	}
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO links(code, url, created_at, expires_at, hit_count) VALUES(?, ?, ?, ?, 0)",
		l.Code, l.URL, time.Now(), expires,
	)
	if err != nil {
		return err
	}
	if id, err := res.LastInsertId(); err == nil {
		l.ID = id
	}
	return nil
}

func scanLink(row *sql.Row) (*repository.Link, error) {
	var l repository.Link
	var expires sql.NullTime
	if err := row.Scan(&l.ID, &l.Code, &l.URL, &l.CreatedAt, &expires, &l.HitCount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if expires.Valid {
		t := expires.Time
		l.ExpiresAt = &t
	}
	return &l, nil
}

func (r *LinkRepo) GetByCode(ctx context.Context, code string) (*repository.Link, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, url, created_at, expires_at, hit_count
           FROM links
          WHERE code = ?
            AND deleted_at IS NULL
          LIMIT 1`, code,
	)
	return scanLink(row)
}

func (r *LinkRepo) GetByURL(ctx context.Context, url string) (*repository.Link, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, code, url, created_at, expires_at, hit_count
           FROM links
          WHERE url = ?
            AND deleted_at IS NULL
       ORDER BY id DESC
          LIMIT 1`, url,
	)
	return scanLink(row)
}

func (r *LinkRepo) IncrementHit(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE links
            SET hit_count = hit_count + 1
          WHERE code = ?`, code,
	)
	return err
}

func (r *LinkRepo) SoftDelete(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE links
            SET deleted_at = CURRENT_TIMESTAMP
          WHERE code = ?
            AND deleted_at IS NULL`, code,
	)
	return err
}

func (r *LinkRepo) PurgeDeleted(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM links
          WHERE deleted_at IS NOT NULL
            AND deleted_at < ?`, cutoff,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
