package http

import (
	"log/slog"
	"net/http"
)

func NewRouter(h *Handlers, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.Health)
	mux.HandleFunc("/shorten", h.Shorten)
	mux.HandleFunc("/r/", h.Redirect)
	mux.HandleFunc("/stats/", h.Stats)
	//mux.HandleFunc("/", h.Redirect)
	mux.HandleFunc("/delete/", h.Delete)

	return WithMiddlewares(mux, RequestLogger(log))
}
