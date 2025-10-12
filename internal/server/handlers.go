package server

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"git.tyss.io/cj3636/dman/internal/logx"
	"git.tyss.io/cj3636/dman/internal/storage"
	"git.tyss.io/cj3636/dman/pkg/model"
)

type cfgUsers interface{ UsersList() []model.UserSpec }

func compareHandler(store storage.Backend, cmp Comparator, cfg cfgUsers, logger *logx.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = cfg // reserved for future use (e.g., validation)
		var req model.CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		includeSame := r.URL.Query().Get("include_same") == "1"
		serverInv, err := buildStoreInventory(store, req.Users)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		changes := cmp.Compare(req, serverInv)
		if includeSame {
			cmap := map[string]model.InventoryItem{}
			for _, it := range req.Inventory {
				cmap[it.User+"::"+it.Path] = it
			}
			smap := map[string]model.InventoryItem{}
			for _, it := range serverInv {
				smap[it.User+"::"+it.Path] = it
			}
			for k, cit := range cmap {
				if sit, ok := smap[k]; ok && sit.Hash == cit.Hash {
					changes = append(changes, model.Change{User: cit.User, Path: cit.Path, Type: model.ChangeSame})
				}
			}
		}
		logger.Info("compare", "changes", len(changes))
		if err := json.NewEncoder(w).Encode(changes); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

func uploadHandler(store storage.Backend, logger *logx.Logger) http.HandlerFunc {
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
		logger.Info("upload", "user", user, "path", p)
		w.WriteHeader(http.StatusNoContent)
	}
}

func downloadHandler(store storage.Backend, logger *logx.Logger) http.HandlerFunc {
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
		logger.Info("download", "user", user, "path", p)
	}
}

// publishHandler accepts a tar stream (application/x-tar) of files named user/relative/path
// and stores them. Responds with JSON summary {"stored":N}.
func publishHandler(store storage.Backend, meta *Meta, logger *logx.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/x-tar") {
			http.Error(w, "expected application/x-tar", http.StatusUnsupportedMediaType)
			return
		}
		var reader io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "bad gzip", 400)
				return
			}
			defer gzr.Close()
			reader = gzr
		}
		tr := tar.NewReader(reader)
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
				code := 500
				if strings.Contains(err.Error(), "too long") {
					code = 400
				}
				http.Error(w, err.Error(), code)
				return
			}
			stored++
		}
		meta.recordPublish()
		logger.Info("publish complete", "stored", stored)
		if err := json.NewEncoder(w).Encode(map[string]int{"stored": stored}); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

// installHandler accepts a CompareRequest JSON body and returns a tar containing the
// files that should be installed locally (ChangeDelete or ChangeModify).
func installHandler(store storage.Backend, cmp Comparator, cfg cfgUsers, meta *Meta, logger *logx.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = cfg
		var req model.CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		serverInv, err := buildStoreInventory(store, req.Users)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		changes := cmp.Compare(req, serverInv)
		accept := r.Header.Get("Accept-Encoding")
		var writer io.Writer = w
		if strings.Contains(accept, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gw := gzip.NewWriter(w)
			defer gw.Close()
			writer = gw
		}
		w.Header().Set("Content-Type", "application/x-tar")
		include := func(ch model.Change) bool { return ch.Type == model.ChangeDelete || ch.Type == model.ChangeModify }
		writeChangesTar(store, changes, writer, include)
		meta.recordInstall()
		logger.Info("install streamed", "files", len(changes))
	}
}

func pruneHandler(store storage.Backend, logger *logx.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Deletes []struct {
				User string `json:"user"`
				Path string `json:"path"`
			} `json:"deletes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		deleted := 0
		for _, d := range body.Deletes {
			if d.User == "" || d.Path == "" {
				continue
			}
			_ = store.Delete(d.User, d.Path)
			deleted++
		}
		if err := json.NewEncoder(w).Encode(map[string]int{"deleted": deleted}); err != nil {
			http.Error(w, err.Error(), 500)
		}
		logger.Info("prune", "deleted", deleted)
	}
}

func buildStoreInventory(store storage.Backend, filterUsers []string) ([]model.InventoryItem, error) {
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
