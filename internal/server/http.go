package server

import (
	"context"
	"net/http"
	"time"
)

type HTTPServer struct {
	srv *http.Server
}

func NewHTTPServer(addr string, handler http.Handler, rt, wt, it time.Duration) *HTTPServer {
	return &HTTPServer{
		srv: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  rt,
			WriteTimeout: wt,
			IdleTimeout:  it,
		},
	}
}

func (s *HTTPServer) Start() error {
	return s.srv.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
