package storage

import (
	"io"
	"os"
)

// mariaBackend is a scaffold that delegates to a disk Store for now.
// Future implementation will persist metadata & blobs in MariaDB.
type mariaBackend struct{ store *Store }

// NewMariaScaffoldBackend creates a MariaDB scaffold backend (disk-backed). Retained for legacy tests; not used by factory.
func NewMariaScaffoldBackend(root string) (Backend, error) {
	st, err := New(root + "/maria")
	if err != nil {
		return nil, err
	}
	return &mariaBackend{store: st}, nil
}

func (m *mariaBackend) Save(user, rel string, r io.Reader) error { return m.store.Save(user, rel, r) }
func (m *mariaBackend) Open(user, rel string) (*os.File, error)  { return m.store.Open(user, rel) }
func (m *mariaBackend) List() ([]string, error)                  { return m.store.List() }
func (m *mariaBackend) Delete(user, rel string) error            { return m.store.Delete(user, rel) }
