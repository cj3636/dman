package storage

import (
	"bytes"
	"os"
	"testing"

	"git.tyss.io/cj3636/dman/internal/config"
)

// Optional integration test for MariaDB; requires DMAN_TEST_MARIA_* environment variables.
// DMAN_TEST_MARIA_ADDR (host:port) or DMAN_TEST_MARIA_SOCKET, DMAN_TEST_MARIA_DB, DMAN_TEST_MARIA_USER, DMAN_TEST_MARIA_PASS
func TestMariaRealBackendOptional(t *testing.T) {
	addr := os.Getenv("DMAN_TEST_MARIA_ADDR")
	socket := os.Getenv("DMAN_TEST_MARIA_SOCKET")
	db := os.Getenv("DMAN_TEST_MARIA_DB")
	user := os.Getenv("DMAN_TEST_MARIA_USER")
	pass := os.Getenv("DMAN_TEST_MARIA_PASS")
	if db == "" || user == "" || (addr == "" && socket == "") {
		t.Skip("maria env vars not set")
	}
	cfg := &config.Config{ServerURL: "http://x", StorageDriver: "mariadb", Users: map[string]config.User{"u": {Home: "/h/"}}, Maria: config.Maria{Addr: addr, Socket: socket, DB: db, User: user, Password: pass}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	b, err := NewBackend(cfg, "data")
	if err != nil {
		t.Fatalf("backend: %v", err)
	}
	if err := b.Save("u", "afile.txt", bytes.NewReader([]byte("maria"))); err != nil {
		t.Fatalf("save: %v", err)
	}
	f, err := b.Open("u", "afile.txt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	buf := make([]byte, 5)
	if _, err := f.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	f.Close()
	if string(buf) != "maria" {
		t.Fatalf("unexpected data %q", string(buf))
	}
}
