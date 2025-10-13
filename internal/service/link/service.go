package link

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"tinyurl/internal/repository"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrAliasBusy = errors.New("alias is already in use")
	ErrExpired   = errors.New("link expired")
)

type Service struct {
	repo repository.LinkRepository
}

func New(repo repository.LinkRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Shorten(ctx context.Context, url, alias string, ttlDays int) (string, *time.Time, error) {
	url = strings.TrimSpace(url)
	alias = strings.TrimSpace(alias)

	if url == "" {
		return "", nil, errors.New("empty url")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	var expiresAt *time.Time
	if ttlDays > 0 {
		t := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour)
		expiresAt = &t
	}

	if alias != "" {
		return s.withAlias(ctx, url, alias, expiresAt)
	}

	return s.withCode(ctx, url, expiresAt)
}

func (s *Service) withCode(ctx context.Context, url string, expiresAt *time.Time) (string, *time.Time, error) {
	if existing, _ := s.repo.GetByURL(ctx, url); existing != nil && !isExpired(existing) {
		return existing.Code, existing.ExpiresAt, nil
	}

	code, err := s.generateCode(ctx)
	if err != nil {
		return "", nil, err
	}

	link := &repository.Link{
		Code:      code,
		URL:       url,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.Create(ctx, link); err != nil {
		return "", nil, err
	}

	return code, expiresAt, nil
}

func (s *Service) withAlias(ctx context.Context, url, alias string, expiresAt *time.Time) (string, *time.Time, error) {
	if err := s.aliasAvailable(ctx, alias); err != nil {
		return "", nil, err
	}

	link := &repository.Link{
		Code:      alias,
		URL:       url,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.Create(ctx, link); err != nil {
		return "", nil, err
	}

	return alias, expiresAt, nil
}

func (s *Service) aliasAvailable(ctx context.Context, alias string) error {
	existing, err := s.repo.GetByCode(ctx, alias)
	if err != nil {
		return err
	}

	if existing == nil {
		return nil
	}

	if !isExpired(existing) {
		return ErrAliasBusy
	}

	return s.repo.SoftDelete(ctx, alias)
}

func (s *Service) generateCode(ctx context.Context) (string, error) {
	for i := 0; i < 6; i++ {
		cand := genCode(7 + i%2)

		got, _ := s.repo.GetByCode(ctx, cand)
		if got == nil || isExpired(got) {
			return cand, nil
		}
	}
	return "", errors.New("cannot generate code")
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", ErrNotFound
	}

	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", err
	}
	if l == nil {
		return "", ErrNotFound
	}
	if isExpired(l) {
		return "", ErrExpired
	}
	_ = s.repo.IncrementHit(ctx, code)
	return l.URL, nil
}

func (s *Service) Stats(ctx context.Context, code string) (*repository.Link, error) {
	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrNotFound
	}
	if isExpired(l) {
		return nil, ErrExpired
	}
	return l, nil
}

func isExpired(l *repository.Link) bool {
	return l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt)
}

func (s *Service) Delete(ctx context.Context, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return ErrNotFound
	}
	l, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if l == nil {
		return ErrNotFound
	}
	if isExpired(l) {
		return ErrExpired
	}
	return s.repo.SoftDelete(ctx, code)
}

func genCode(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	s := base64.RawURLEncoding.EncodeToString(b)
	if len(s) > n {
		s = s[:n]
	}
	return s
}
