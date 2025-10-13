package tests

import (
	"context"
	"testing"
	"time"

	"tinyurl/internal/repository"
)

func TestShorten_And_Resolve(t *testing.T) {
	d := NewTestDeps(t)
	ctx := context.Background()

	code, _, err := d.Svc.Shorten(ctx, "https://example.com", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if code == "" {
		t.Fatal("empty code")
	}
	url, err := d.Svc.Resolve(ctx, code)
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://example.com" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestShorten_WithAlias(t *testing.T) {
	d := NewTestDeps(t)
	ctx := context.Background()

	code, _, err := d.Svc.Shorten(ctx, "https://golang.org", "go", 0)
	if err != nil {
		t.Fatal(err)
	}
	if code != "go" {
		t.Fatalf("alias mismatch: %s", code)
	}

	if _, _, err := d.Svc.Shorten(ctx, "https://golang.org/doc", "go", 0); err == nil {
		t.Fatal("expected alias conflict, got nil")
	}
}

func TestResolve_Expired(t *testing.T) {
	d := NewTestDeps(t)
	ctx := context.Background()

	past := time.Now().Add(-24 * time.Hour)
	l := repository.Link{Code: "old", URL: "https://old.example", ExpiresAt: &past}
	if err := d.Rep.Create(ctx, &l); err != nil {
		t.Fatal(err)
	}

	if _, err := d.Svc.Resolve(ctx, "old"); err == nil {
		t.Fatal("expected not found for expired link")
	}
}
