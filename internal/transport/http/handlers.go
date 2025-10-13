package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"tinyurl/internal/service/link"
)

type Handlers struct {
	svc     *link.Service
	log     *slog.Logger
	baseURL string
}

func NewHandlers(svc *link.Service, log *slog.Logger, baseURL string) *Handlers {
	return &Handlers{svc: svc, log: log, baseURL: strings.TrimRight(baseURL, "/")}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handlers) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req ShortenRequest
	if err := dec.Decode(&req); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.Is(err, io.EOF):
			http.Error(w, "empty body", http.StatusBadRequest)

		case errors.As(err, &syntaxErr):
			msg := fmt.Sprintf("malformed JSON (syntax error at pos %d)", syntaxErr.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeErr):
			msg := fmt.Sprintf("invalid value for field %q (at pos %d)", unmarshalTypeErr.Field, unmarshalTypeErr.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), `json: unknown field `)
			msg := fmt.Sprintf("unknown field %s", field)
			http.Error(w, msg, http.StatusBadRequest)

		default:
			http.Error(w, "invalid request format", http.StatusBadRequest)
			h.log.Warn("shorten: decode error", "err", err)
		}

		h.log.Warn("shorten: bad JSON", "err", err)
		return
	}

	if dec.More() {
		http.Error(w, "only one JSON object allowed", http.StatusBadRequest)
		return
	}
	code, expiresAt, err := h.svc.Shorten(r.Context(), req.URL, req.Alias, req.TTLDays)
	if err != nil {
		switch {
		case errors.Is(err, link.ErrAliasBusy):
			http.Error(w, "alias is already in use", http.StatusConflict)
		default:
			http.Error(w, "cannot shorten", http.StatusBadRequest)
		}
		h.log.Warn("shorten failed", "err", err)
		return
	}
	short := h.baseURL + "/r/" + code
	resp := ShortenResponse{Code: code, ShortURL: short}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

	if expiresAt != nil {
		h.log.Info("shortened", "code", code, "url", req.URL, "expires_at", expiresAt.Format(time.RFC3339))
	} else {
		h.log.Info("shortened", "code", code, "url", req.URL)
	}
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/r/")

	h.log.Info("redirect request", "path", r.URL.Path, "code", code)

	if code == "" || strings.Contains(code, "/") {
		http.NotFound(w, r)
		h.log.Warn("invalid code format", "code", code)
		return
	}

	url, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		h.log.Warn("code not found", "code", code, "error", err)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
	h.log.Info("redirect success", "code", code, "to", url)
}

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/stats/")
	if code == "" || strings.Contains(code, "/") {
		http.NotFound(w, r)
		return
	}
	l, err := h.svc.Stats(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var expiresStr *string
	if l.ExpiresAt != nil {
		s := l.ExpiresAt.UTC().Format(time.RFC3339)
		expiresStr = &s
	}

	resp := StatsResponse{
		URL:       l.URL,
		CreatedAt: l.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: expiresStr,
		HitCount:  l.HitCount,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/delete/")
	if code == "" {
		http.Error(w, "code required", http.StatusBadRequest)
		return
	}

	err := h.svc.Delete(r.Context(), code)
	if err != nil {
		if errors.Is(err, link.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "cannot delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
