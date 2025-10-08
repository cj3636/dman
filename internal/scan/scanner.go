package scan

import (
	"crypto/sha256"
	"encoding/hex"
	"git.tyss.io/cj3636/dman/pkg/model"
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
		for _, inc := range u.Include {
			if strings.HasSuffix(inc, "/") { // directory
				root := filepath.Join(u.Home, inc)
				filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if d.IsDir() {
						return nil
					}
					rel, _ := filepath.Rel(u.Home, path)
					item, err := fileItem(u.Name, path, rel)
					if err == nil {
						out = append(out, item)
					}
					return nil
				})
			} else {
				abs := filepath.Join(u.Home, inc)
				fi, err := os.Stat(abs)
				if err != nil || fi.IsDir() {
					continue
				}
				item, err := fileItem(u.Name, abs, inc)
				if err == nil {
					out = append(out, item)
				}
			}
		}
	}
	return out, nil
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
