package app

import (
	"context"
	"log/slog"
	"time"

	"tinyurl/internal/config"
	"tinyurl/internal/logger"
	"tinyurl/internal/repository"
	"tinyurl/internal/server"
	"tinyurl/internal/service/link"
	"tinyurl/internal/storage/sqlite"
	sqlrepo "tinyurl/internal/storage/sqlite/repository"
	thttp "tinyurl/internal/transport/http"
)

type App struct {
	Log     *slog.Logger
	Server  *server.HTTPServer
	DBClose func() error
}

func Build(ctx context.Context, cfg config.Config) (*App, error) {
	log := logger.New(cfg.Log.Level)

	db, err := sqlite.Open(cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	if err := sqlite.Migrate(ctx, db.DB); err != nil {
		return nil, err
	}

	var linkRepo repository.LinkRepository = sqlrepo.NewLinkRepo(db.DB)
	linkSvc := link.New(linkRepo)

	h := thttp.NewHandlers(linkSvc, log, cfg.HTTP.BaseURL)
	router := thttp.NewRouter(h, log)

	s := server.NewHTTPServer(
		cfg.HTTP.Address,
		router,
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
	)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				repo := sqlrepo.NewLinkRepo(db.DB)
				n, err := repo.PurgeDeleted(context.Background(), 24*time.Hour)
				if err != nil {
					log.Error("purge deleted links", "err", err)
				} else if n > 0 {
					log.Info("purged deleted links", "count", n)
				}
			}
		}
	}()

	return &App{
		Log:     log,
		Server:  s,
		DBClose: db.Close,
	}, nil
}
