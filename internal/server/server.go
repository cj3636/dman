package server

import (
	"encoding/json"
	"net/http"
	"time"

	"git.tyss.io/cj3636/dman/internal/auth"
	"git.tyss.io/cj3636/dman/internal/buildinfo"
	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/logx"
	"git.tyss.io/cj3636/dman/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(addr string, cfg *config.Config, logger *logx.Logger) (*http.Server, error) {
	store, err := storage.NewBackend(cfg, "data")
	if err != nil {
		return nil, err
	}
	meta, err := loadMeta("data")
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = logx.New()
	}
	h := newHandler(cfg, store, meta, logger)
	return &http.Server{Addr: addr, Handler: h}, nil
}

func newHandler(cfg *config.Config, store storage.Backend, meta *Meta, logger *logx.Logger) http.Handler {
	cmp := diffComparator()
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, requestLogger(logger))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"ok": true, "version": buildinfo.Version, "build_time": buildinfo.BuildTime, "commit": buildinfo.Commit, "server_time": time.Now().UTC().Format(time.RFC3339)}
		_ = json.NewEncoder(w).Encode(resp)
	})
	r.Group(func(pr chi.Router) {
		pr.Use(auth.Bearer(cfg.AuthToken))
		pr.Post("/compare", compareHandler(store, cmp, cfg, logger))
		pr.Post("/publish", publishHandler(store, meta, logger))
		pr.Post("/install", installHandler(store, cmp, cfg, meta, logger))
		pr.Post("/prune", pruneHandler(store, logger))
		pr.Put("/upload", uploadHandler(store, logger))
		pr.Get("/download", downloadHandler(store, logger))
		pr.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			st, err := buildStatus(store, meta)
			if err != nil {
				logger.Error("status error", "err", err)
				http.Error(w, err.Error(), 500)
				return
			}
			_ = json.NewEncoder(w).Encode(st)
		})
	})
	return r
}

func requestLogger(l *logx.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			l.Info("request", "method", r.Method, "path", r.URL.Path, "dur_ms", time.Since(start).Milliseconds())
		})
	}
}

func diffComparator() Comparator { return newComparator() }
