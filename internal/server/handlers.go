package server

import (
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
