package storage

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxPathLen = 4096

// Store manages versionless file blobs under a root directory organized by user.
// Layout: root/<user>/<relative file path>
// Only simple traversal protections; not intended for untrusted remote user input beyond controlled API layer.
// Not concurrency-optimized; sufficient for small LAN usage.
type Store struct{ root string }

// New ensures the root directory exists and returns a Store.
func New(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("empty root")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Store{root: root}, nil
}

// sanitize validates and normalizes a relative path (forward slashes, rejects traversal/absolute).
func (s *Store) sanitize(rel string) (string, error) {
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "./")
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
	return rel, nil
}

// Save writes content for a given user & relative path atomically (simple overwrite behavior).
func (s *Store) Save(user, rel string, r io.Reader) error {
	rel, err := s.sanitize(rel)
	if err != nil {
		return err
	}
	absDir := filepath.Join(s.root, user, filepath.Dir(filepath.FromSlash(rel)))
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return err
	}
	finalPath := filepath.Join(s.root, user, filepath.FromSlash(rel))
	tmp, err := os.CreateTemp(absDir, ".dman-*.")
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(tmp, r)
	// Attempt to fsync to reduce risk of partial writes on power loss (best-effort)
	_ = tmp.Sync()
	closeErr := tmp.Close()
	if copyErr != nil {
		os.Remove(tmp.Name())
		return copyErr
	}
	if closeErr != nil {
		os.Remove(tmp.Name())
		return closeErr
	}
	if err := os.Rename(tmp.Name(), finalPath); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}

// Open opens a stored file for reading.
func (s *Store) Open(user, rel string) (*os.File, error) {
	rel, err := s.sanitize(rel)
	if err != nil {
		return nil, err
	}
	abs := filepath.Join(s.root, user, filepath.FromSlash(rel))
	return os.Open(abs)
}

// List returns all stored files as paths in form "user/relpath" using forward slashes.
func (s *Store) List() ([]string, error) {
	var out []string
	err := filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(s.root, path)
		if rel == "_meta.json" { // skip meta file if placed at root
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out, err
}

// Backup writes a tar archive of all files to w.
func (s *Store) Backup(w io.Writer) error {
	files, err := s.List()
	if err != nil {
		return err
	}
	tw := tar.NewWriter(w)
	defer tw.Close()
	for _, rel := range files {
		abs := filepath.Join(s.root, filepath.FromSlash(rel))
		fi, err := os.Stat(abs)
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		hdr.Name = rel // ensure forward slash form
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		f, err := os.Open(abs)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

// Restore reads a tar archive from r and recreates files (overwrites existing).
func (s *Store) Restore(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		clean, err := s.sanitize(hdr.Name)
		if err != nil {
			return err
		}
		abs := filepath.Join(s.root, filepath.FromSlash(clean))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		f, err := os.Create(abs)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
		// attempt to preserve mod time
		_ = os.Chtimes(abs, time.Now(), hdr.ModTime)
	}
}

// Delete removes a stored file for a user.
func (s *Store) Delete(user, rel string) error {
	rel, err := s.sanitize(rel)
	if err != nil {
		return err
	}
	abs := filepath.Join(s.root, user, filepath.FromSlash(rel))
	return os.Remove(abs)
}
