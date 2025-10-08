package server

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"git.tyss.io/cj3636/dman/internal/storage"
	"git.tyss.io/cj3636/dman/pkg/model"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"ok":true}`)
	}
}

func statusHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, _ := store.List()
		resp := map[string]any{"files": len(files)}
		json.NewEncoder(w).Encode(resp)
	}
}

type cfgUsers interface{ UsersList() []model.UserSpec }

func compareHandler(store *storage.Store, cmp Comparator, cfg cfgUsers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		serverInv, _ := buildStoreInventory(store, req.Users)
		changes := cmp.Compare(req, serverInv)
		json.NewEncoder(w).Encode(changes)
	}
}

func uploadHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		p := r.URL.Query().Get("path")
		if user == "" || p == "" {
			http.Error(w, "missing user/path", http.StatusBadRequest)
			return
		}
		if err := store.Save(user, filepath.Clean(p), r.Body); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func downloadHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		p := r.URL.Query().Get("path")
		if user == "" || p == "" {
			http.Error(w, "missing user/path", http.StatusBadRequest)
			return
		}
		rc, err := store.Open(user, filepath.Clean(p))
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		defer rc.Close()
		io.Copy(w, rc)
	}
}

// publishHandler accepts a tar stream (application/x-tar) of files named user/relative/path
// and stores them. Responds with JSON summary {"stored":N}.
func publishHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/x-tar") { // allow parameters
			http.Error(w, "expected application/x-tar", http.StatusUnsupportedMediaType)
			return
		}
		tr := tar.NewReader(r.Body)
		stored := 0
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if hdr.FileInfo().IsDir() {
				continue
			}
			name := filepath.ToSlash(hdr.Name)
			parts := strings.SplitN(name, "/", 2)
			if len(parts) != 2 {
				http.Error(w, "invalid entry name", 400)
				return
			}
			user, rel := parts[0], parts[1]
			if err := store.Save(user, rel, tr); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			stored++
		}
		json.NewEncoder(w).Encode(map[string]int{"stored": stored})
	}
}

// installHandler accepts a CompareRequest JSON body and returns a tar containing the
// files that should be installed locally (ChangeDelete or ChangeModify).
func installHandler(store *storage.Store, cmp Comparator, cfg cfgUsers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		serverInv, _ := buildStoreInventory(store, req.Users)
		changes := cmp.Compare(req, serverInv)
		w.Header().Set("Content-Type", "application/x-tar")
		tw := tar.NewWriter(w)
		defer tw.Close()
		for _, ch := range changes {
			if ch.Type != model.ChangeDelete && ch.Type != model.ChangeModify {
				continue
			}
			f, err := store.Open(ch.User, ch.Path)
			if err != nil {
				continue
			}
			fi, _ := f.Stat()
			hdr, _ := tar.FileInfoHeader(fi, "")
			hdr.Name = ch.User + "/" + ch.Path
			if err := tw.WriteHeader(hdr); err != nil {
				f.Close()
				continue
			}
			io.Copy(tw, f)
			f.Close()
		}
	}
}

func buildStoreInventory(store *storage.Store, filterUsers []string) ([]model.InventoryItem, error) {
	files, err := store.List()
	if err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{}
	if len(filterUsers) > 0 {
		for _, u := range filterUsers {
			allowed[u] = struct{}{}
		}
	}
	var inv []model.InventoryItem
	for _, rel := range files { // rel: user/path
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) != 2 {
			continue
		}
		user, p := parts[0], parts[1]
		if len(allowed) > 0 {
			if _, ok := allowed[user]; !ok {
				continue
			}
		}
		f, err := store.Open(user, p)
		if err != nil {
			continue
		}
		h := sha256.New()
		sz, _ := io.Copy(h, f)
		fi, _ := f.Stat()
		f.Close()
		inv = append(inv, model.InventoryItem{User: user, Path: p, Size: sz, MTime: fi.ModTime().Unix(), Hash: hex.EncodeToString(h.Sum(nil)), IsDir: false})
	}
	return inv, nil
}
