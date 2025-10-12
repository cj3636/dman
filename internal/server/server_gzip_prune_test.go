package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/logx"
	"git.tyss.io/cj3636/dman/internal/storage"
	"git.tyss.io/cj3636/dman/pkg/model"
)

// Test bulk publish with gzip and install with gzip response.
func TestGzipPublishInstall(t *testing.T) {
	cfg := &config.Config{AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/"}}}
	store, _ := storage.New(t.TempDir())
	meta, _ := loadMeta(t.TempDir())
	logger := logx.New()
	h := newHandler(cfg, store, meta, logger)
	ts := httptest.NewServer(h)
	defer ts.Close()

	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	content := []byte("data")
	tw.WriteHeader(&tar.Header{Name: "u/file.txt", Mode: 0o644, Size: int64(len(content))})
	tw.Write(content)
	tw.Close()
	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	gw.Write(tarBuf.Bytes())
	gw.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(gzBuf.Bytes()))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/x-tar")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("publish gzip status %d", resp.StatusCode)
	}
	resp.Body.Close()

	comp := model.CompareRequest{Users: []string{"u"}, Inventory: []model.InventoryItem{}}
	b, _ := json.Marshal(comp)
	instReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/install", bytes.NewReader(b))
	instReq.Header.Set("Authorization", "Bearer tok")
	instReq.Header.Set("Content-Type", "application/json")
	instReq.Header.Set("Accept-Encoding", "gzip")
	instResp, err := http.DefaultClient.Do(instReq)
	if err != nil {
		t.Fatal(err)
	}
	if instResp.StatusCode != 200 {
		t.Fatalf("install gzip status %d", instResp.StatusCode)
	}
	if instResp.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response")
	}
	gzr, err := gzip.NewReader(instResp.Body)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gzr)
	if _, err := tr.Next(); err != nil {
		t.Fatalf("expected tar entry: %v", err)
	}
	instResp.Body.Close()
}

func TestPruneEndpoint(t *testing.T) {
	cfg := &config.Config{AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/"}}}
	storeDir := t.TempDir()
	store, _ := storage.New(storeDir)
	meta, _ := loadMeta(storeDir)
	logger := logx.New()
	h := newHandler(cfg, store, meta, logger)
	ts := httptest.NewServer(h)
	defer ts.Close()
	// upload one file via publish
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	content := []byte("hello")
	tw.WriteHeader(&tar.Header{Name: "u/obsolete.txt", Mode: 0o644, Size: int64(len(content))})
	tw.Write(content)
	tw.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(tarBuf.Bytes()))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/x-tar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("publish failed %v status %d", err, resp.StatusCode)
	}
	resp.Body.Close()
	// prune
	body := []byte(`{"deletes":[{"user":"u","path":"obsolete.txt"}]}`)
	pr, _ := http.NewRequest(http.MethodPost, ts.URL+"/prune", bytes.NewReader(body))
	pr.Header.Set("Authorization", "Bearer tok")
	pr.Header.Set("Content-Type", "application/json")
	prResp, err := http.DefaultClient.Do(pr)
	if err != nil || prResp.StatusCode != 200 {
		t.Fatalf("prune failed %v status %d", err, prResp.StatusCode)
	}
	prResp.Body.Close()
	// ensure file deleted
	files, _ := store.List()
	for _, f := range files {
		if strings.Contains(f, "obsolete.txt") {
			t.Fatalf("file not pruned")
		}
	}
}
