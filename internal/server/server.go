package server

import (
	"git.tyss.io/cj3636/dman/internal/auth"
	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func New(addr string, cfg *config.Config) (*http.Server, error) {
	store, err := storage.New("data")
	if err != nil {
		return nil, err
	}
	cmp := diffComparator()
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Get("/health", healthHandler())
	r.Group(func(pr chi.Router) {
		pr.Use(auth.Bearer(cfg.AuthToken))
		pr.Get("/status", statusHandler(store))
		pr.Post("/compare", compareHandler(store, cmp, cfg))
		pr.Put("/upload", uploadHandler(store))
		pr.Get("/download", downloadHandler(store))
	})
	return &http.Server{Addr: addr, Handler: r}, nil
}

// small indirection avoids import cycle
func diffComparator() Comparator { return newComparator() }
