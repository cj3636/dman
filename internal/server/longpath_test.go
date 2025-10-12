package server

import (
	"archive/tar"
	"bytes"
	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/logx"
	"git.tyss.io/cj3636/dman/internal/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishRejectsLongPath(t *testing.T) {
	cfg := &config.Config{AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/"}}}
	store, _ := storage.New(t.TempDir())
	meta, _ := loadMeta(t.TempDir())
	logger := logx.New()
	h := newHandler(cfg, store, meta, logger)
	ts := httptest.NewServer(h)
	defer ts.Close()

	longName := strings.Repeat("a", storage.MaxPathLen+1)
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	data := []byte("x")
	tw.WriteHeader(&tar.Header{Name: "u/" + longName, Mode: 0o644, Size: int64(len(data))})
	tw.Write(data)
	tw.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Type", "application/x-tar")
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}
