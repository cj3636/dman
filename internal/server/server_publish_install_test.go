package server

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/storage"
	"git.tyss.io/cj3636/dman/pkg/model"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublishAndInstallEndpoints(t *testing.T) {
	cfg := &config.Config{AuthToken: "tok", Users: map[string]config.User{"u": {Home: t.TempDir() + "/", Include: []string{"file.txt"}}}}
	store, _ := storage.New(t.TempDir())
	h := newHandler(cfg, store)
	ts := httptest.NewServer(h)
	defer ts.Close()

	// Build tar with one file user u/file.txt
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	{ // single file entry
		data := []byte("hello world")
		hdr := &tar.Header{Name: "u/file.txt", Mode: 0o644, Size: int64(len(data))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/publish", bytes.NewReader(tarBuf.Bytes()))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/x-tar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("publish status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Now request install with empty inventory (should get the file)
	comp := model.CompareRequest{Users: []string{"u"}, Inventory: []model.InventoryItem{}}
	b, _ := json.Marshal(comp)
	instReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/install", bytes.NewReader(b))
	instReq.Header.Set("Authorization", "Bearer tok")
	instReq.Header.Set("Content-Type", "application/json")
	instResp, err := http.DefaultClient.Do(instReq)
	if err != nil {
		t.Fatal(err)
	}
	if instResp.StatusCode != 200 {
		t.Fatalf("install status %d", instResp.StatusCode)
	}
	// parse tar
	tr := tar.NewReader(instResp.Body)
	entries := 0
	for {
		_, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar error: %v", err)
		}
		entries++
	}
	instResp.Body.Close()
	if entries != 1 {
		t.Fatalf("expected 1 entry got %d", entries)
	}
}
