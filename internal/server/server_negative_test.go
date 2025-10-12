package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/logx"
	"git.tyss.io/cj3636/dman/internal/storage"
)

// Test invalid gzip Content-Encoding on /publish
func TestPublishInvalidGzip(t *testing.T) {
	cfg := &config.Config{ServerURL: "http://x", AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/"}}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	store, _ := storage.New(t.TempDir())
	meta, _ := loadMeta(t.TempDir())
	logger := logx.New()
	h := newHandler(cfg, store, meta, logger)
	ts := httptest.NewServer(h)
	defer ts.Close()

	bad := []byte("not-a-gzip-stream")
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(bad))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/x-tar")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

// Test invalid tar (not a tar archive) without gzip
func TestPublishInvalidTar(t *testing.T) {
	cfg := &config.Config{ServerURL: "http://x", AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/"}}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	store, _ := storage.New(t.TempDir())
	meta, _ := loadMeta(t.TempDir())
	logger := logx.New()
	h := newHandler(cfg, store, meta, logger)
	ts := httptest.NewServer(h)
	defer ts.Close()

	bad := []byte("not-a-tar")
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(bad))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/x-tar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}
