package server

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/logx"
	"git.tyss.io/cj3636/dman/internal/storage"
)

func TestStatusMetricsAfterPublish(t *testing.T) {
	cfg := &config.Config{AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/"}}}
	store, _ := storage.New(t.TempDir())
	meta, _ := loadMeta(t.TempDir())
	logger := logx.New()
	h := newHandler(cfg, store, meta, logger)
	ts := httptest.NewServer(h)
	defer ts.Close()

	// publish empty tar (no files) still increments metric
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	tw.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(tarBuf.Bytes()))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/x-tar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("publish %v status %d", err, resp.StatusCode)
	}
	resp.Body.Close()

	// status
	sreq, _ := http.NewRequest(http.MethodGet, ts.URL+"/status", nil)
	sreq.Header.Set("Authorization", "Bearer tok")
	sresp, err := http.DefaultClient.Do(sreq)
	if err != nil {
		t.Fatal(err)
	}
	if sresp.StatusCode != 200 {
		t.Fatalf("status code %d", sresp.StatusCode)
	}
	var body map[string]any
	json.NewDecoder(sresp.Body).Decode(&body)
	sresp.Body.Close()
	metrics, ok := body["metrics"].(map[string]any)
	if !ok {
		t.Fatalf("missing metrics")
	}
	if metrics["publish_requests"].(float64) < 1 {
		t.Fatalf("expected publish_requests >=1 got %v", metrics["publish_requests"])
	}
	if body["last_publish"].(string) == "" {
		t.Fatalf("expected last_publish set")
	}
	// sanity timestamp parse
	if _, err := time.Parse(time.RFC3339, body["last_publish"].(string)); err != nil {
		t.Fatalf("bad last_publish timestamp")
	}
}
