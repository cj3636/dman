package scan

import (
	"crypto/sha256"
	"encoding/hex"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/bmatcuk/doublestar/v4"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type scanner struct{}

func New() Scanner { return &scanner{} }

type Scanner interface {
	InventoryFor(users []model.UserSpec) ([]model.InventoryItem, error)
}

func (s *scanner) InventoryFor(users []model.UserSpec) ([]model.InventoryItem, error) {
	var out []model.InventoryItem
	for _, u := range users {
		includePatterns, excludePatterns := splitPatterns(u.Track)
		seen := map[string]struct{}{}

		for _, pattern := range includePatterns {
			matches := expandPattern(u.Home, pattern)
			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					continue
				}
				if info.IsDir() {
					filepath.WalkDir(match, func(path string, d os.DirEntry, err error) error {
						if err != nil {
							return nil
						}
						rel := relativePath(u.Home, path)
						if rel == "" {
							return nil
						}
						if d.IsDir() {
							if isExcluded(rel, path, excludePatterns) {
								return filepath.SkipDir
							}
							return nil
						}
						if isExcluded(rel, path, excludePatterns) {
							return nil
						}
						if _, exists := seen[path]; exists {
							return nil
						}
						seen[path] = struct{}{}
						item, err := fileItem(u.Name, path, rel)
						if err == nil {
							out = append(out, item)
						}
						return nil
					})
					continue
				}

				rel := relativePath(u.Home, match)
				if rel == "" {
					continue
				}
				if isExcluded(rel, match, excludePatterns) {
					continue
				}
				if _, exists := seen[match]; exists {
					continue
				}
				seen[match] = struct{}{}
				item, err := fileItem(u.Name, match, rel)
				if err == nil {
					out = append(out, item)
				}
			}
		}
	}
	return out, nil
}

func splitPatterns(patterns []string) (includes []string, excludes []string) {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "!") {
			excludes = append(excludes, strings.TrimPrefix(p, "!"))
			continue
		}
		includes = append(includes, p)
	}
	return includes, excludes
}

func hasGlob(p string) bool {
	return strings.ContainsAny(p, "*?[{")
}

func expandPattern(home, pattern string) []string {
	pat := filepath.FromSlash(pattern)
	if !filepath.IsAbs(pat) {
		pat = filepath.Join(home, pat)
	}
	pat = filepath.Clean(pat)

	var matches []string
	if hasGlob(pat) {
		globs, err := doublestar.FilepathGlob(pat)
		if err == nil {
			matches = append(matches, globs...)
		}
	} else {
		matches = append(matches, pat)
	}
	return matches
}

func isExcluded(rel, abs string, excludes []string) bool {
	rel = filepath.ToSlash(rel)
	abs = filepath.ToSlash(abs)
	for _, ex := range excludes {
		ex = strings.TrimSpace(ex)
		if ex == "" {
			continue
		}
		target := rel
		pattern := ex
		if filepath.IsAbs(ex) {
			target = abs
		}
		if ok, _ := doublestar.PathMatch(filepath.ToSlash(pattern), target); ok {
			return true
		}
	}
	return false
}

func relativePath(home, path string) string {
	rel, err := filepath.Rel(home, path)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

func fileItem(user, abs, rel string) (model.InventoryItem, error) {
	f, err := os.Open(abs)
	if err != nil {
		return model.InventoryItem{}, err
	}
	defer f.Close()
	h := sha256.New()
	sz, _ := io.Copy(h, f)
	fi, _ := f.Stat()
	return model.InventoryItem{User: user, Path: filepath.ToSlash(rel), Size: sz, MTime: fi.ModTime().Unix(), Hash: hex.EncodeToString(h.Sum(nil)), IsDir: false}, nil
}
