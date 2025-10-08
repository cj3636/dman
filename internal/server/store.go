package server

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Store struct{ root string }

func NewStore(root string) *Store { os.MkdirAll(root, 0o755); return &Store{root: root} }

func (s *Store) clean(rel string) (string, error) {
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "./")
	if rel == "" {
		return "", errors.New("empty path")
	}
	if strings.HasPrefix(rel, "/") {
		return "", errors.New("absolute path rejected")
	}
	if strings.Contains(rel, "..") {
		return "", errors.New("path traversal rejected")
	}
	return rel, nil
}

func (s *Store) path(user, rel string) (string, error) {
	cr, err := s.clean(rel)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.root, user, cr), nil
}

func (s *Store) Save(user, rel string, r io.Reader) error {
	p, err := s.path(user, rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *Store) Open(user, rel string) (*os.File, error) {
	p, err := s.path(user, rel)
	if err != nil {
		return nil, err
	}
	return os.Open(p)
}

func (s *Store) Inventory() []string {
	var out []string
	filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out
}
