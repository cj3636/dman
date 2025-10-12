package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// redisMemBackend is an in-memory scaffold simulating a future Redis implementation.
// Keys are stored as user/rel with forward slashes. Contents kept in memory.
// Open materializes a temporary file to satisfy the Backend interface.
type redisMemBackend struct {
	root string
	mu   sync.RWMutex
	data map[string][]byte
}

// NewRedisMemBackend creates the in-memory redis backend (driver redis-mem).
func NewRedisMemBackend(root string) (Backend, error) {
	if root == "" {
		root = "redis"
	}
	return &redisMemBackend{root: root, data: map[string][]byte{}}, nil
}

func (r *redisMemBackend) sanitize(user, rel string) (string, error) {
	if user == "" {
		return "", errors.New("empty user")
	}
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if rel == "" {
		return "", errors.New("empty path")
	}
	if len(rel) > MaxPathLen {
		return "", errors.New("path too long")
	}
	if strings.HasPrefix(rel, "/") {
		return "", errors.New("absolute path disallowed")
	}
	if strings.Contains(rel, "..") {
		return "", errors.New("path traversal disallowed")
	}
	return user + "/" + rel, nil
}

func (r *redisMemBackend) Save(user, rel string, rd io.Reader) error {
	key, err := r.sanitize(user, rel)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(rd)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.data[key] = b
	r.mu.Unlock()
	return nil
}

func (r *redisMemBackend) Open(user, rel string) (*os.File, error) {
	key, err := r.sanitize(user, rel)
	if err != nil {
		return nil, err
	}
	r.mu.RLock()
	b, ok := r.data[key]
	r.mu.RUnlock()
	if !ok {
		return nil, os.ErrNotExist
	}
	f, err := os.CreateTemp("", ".dman-redis-*.")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func (r *redisMemBackend) List() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.data))
	for k := range r.data {
		out = append(out, k)
	}
	return out, nil
}

func (r *redisMemBackend) Delete(user, rel string) error {
	key, err := r.sanitize(user, rel)
	if err != nil {
		return err
	}
	r.mu.Lock()
	delete(r.data, key)
	r.mu.Unlock()
	return nil
}
