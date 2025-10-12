package vcs

import "io"

// Repository defines a future version-control-like abstraction to track versions of stored files.
type Repository interface {
	Commit(user, path string, r io.Reader) (revision string, err error)
	Log(user, path string, limit int) ([]Revision, error)
	Checkout(user, path, revision string) (io.ReadCloser, error)
}

type Revision struct {
	ID      string
	TimeISO string
	Size    int64
	Hash    string
	Message string
}

// NoopRepo is a placeholder implementation that does nothing (for current disk store usage).
type NoopRepo struct{}

func (n *NoopRepo) Commit(user, path string, r io.Reader) (string, error)       { return "", nil }
func (n *NoopRepo) Log(user, path string, limit int) ([]Revision, error)        { return nil, nil }
func (n *NoopRepo) Checkout(user, path, revision string) (io.ReadCloser, error) { return nil, nil }
