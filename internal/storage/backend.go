package storage

import (
	"errors"
	"io"
	"os"

	"git.tyss.io/cj3636/dman/internal/config"
)

// Backend describes the minimal storage operations required by server handlers.
type Backend interface {
	Save(user, rel string, r io.Reader) error
	Open(user, rel string) (*os.File, error)
	List() ([]string, error)
	Delete(user, rel string) error
}

// NewBackend constructs a storage backend based on configuration.
// root is the data directory base (used for disk & maria scaffolds; ignored for redis).
func NewBackend(cfg *config.Config, root string) (Backend, error) {
	driver := cfg.StorageDriver
	switch driver {
	case "", "disk":
		return New(root)
	case "redis":
		return NewRedisBackend(cfg)
	case "redis-mem":
		return NewRedisMemBackend(root)
	case "maria", "mariadb", "mysql":
		if cfg.Maria.DB == "" || cfg.Maria.User == "" { // scaffold fallback for incomplete config
			return NewMariaScaffoldBackend(root)
		}
		return NewMariaRealBackend(cfg)
	default:
		return nil, errors.New("unknown storage driver: " + driver)
	}
}
