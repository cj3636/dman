package storage

import (
	"bytes"
	"os"
	"testing"
	"time"

	"git.tyss.io/cj3636/dman/internal/config"
)

// Integration test (optional) requires a running redis if DMAN_TEST_REDIS_ADDR is set.
func TestRedisBackendIntegrationOptional(t *testing.T) {
	addr := os.Getenv("DMAN_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("DMAN_TEST_REDIS_ADDR not set; skipping real redis test")
	}
	cfg := &config.Config{ServerURL: "http://x", StorageDriver: "redis", Users: map[string]config.User{"u": {Home: "/h/"}}, Redis: config.Redis{Addr: addr}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	b, err := NewBackend(cfg, "data")
	if err != nil {
		t.Fatalf("backend: %v", err)
	}
	if err := b.Save("u", "file.txt", bytes.NewReader([]byte("hello"))); err != nil {
		t.Fatalf("save: %v", err)
	}
	f, err := b.Open("u", "file.txt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	buf := make([]byte, 5)
	if n, err := f.Read(buf); err != nil || n != 5 {
		t.Fatalf("read n=%d err=%v", n, err)
	}
	f.Close()
	if string(buf) != "hello" {
		t.Fatalf("unexpected %q", string(buf))
	}
	list, err := b.List()
	if err != nil || len(list) == 0 {
		t.Fatalf("list err=%v list=%v", err, list)
	}
	if err := b.Delete("u", "file.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// small delay to ensure delete propagation (not usually needed)
	time.Sleep(50 * time.Millisecond)
}
