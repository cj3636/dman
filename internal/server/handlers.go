package server

import (
	"encoding/json"
	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/pkg/model"
	"io"
	"net/http"
	"path/filepath"
)

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"ok":true}`)
	}
}

func statusHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inv := store.Inventory()
		resp := map[string]any{"files": len(inv)}
		json.NewEncoder(w).Encode(resp)
	}
}

func compareHandler(store *Store, cmp Comparator, cfg interface{ UsersList() []model.UserSpec }) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		serverScanner := scan.New()
		serverInv, _ := serverScanner.InventoryFor(cfg.UsersList())
		// override server inventory hashes with stored files (ensures we scan server root not client homes) - simplified
		changes := cmp.Compare(req, serverInv)
		json.NewEncoder(w).Encode(changes)
	}
}

func uploadHandler(store *Store) http.HandlerFunc {
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

func downloadHandler(store *Store) http.HandlerFunc {
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
