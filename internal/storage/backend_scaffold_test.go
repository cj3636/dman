package storage

import (
	"bytes"
	"testing"

	"git.tyss.io/cj3636/dman/internal/config"
)

func TestRedisMemBackend(t *testing.T) {
	cfg := &config.Config{StorageDriver: "redis-mem"}
	b, err := NewBackend(cfg, t.TempDir())
	if err != nil {
		t.Fatalf("new redis-mem backend: %v", err)
	}
	if err := b.Save("user1", "file.txt", bytes.NewReader([]byte("hello"))); err != nil {
		t.Fatalf("save: %v", err)
	}
	f, err := b.Open("user1", "file.txt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	data := make([]byte, 5)
	if _, err := f.Read(data); err != nil {
		t.Fatalf("read: %v", err)
	}
	f.Close()
	if string(data) != "hello" {
		t.Fatalf("unexpected data %q", string(data))
	}
	list, err := b.List()
	if err != nil || len(list) != 1 || list[0] != "user1/file.txt" {
		t.Fatalf("list mismatch: %#v %v", list, err)
	}
	if err := b.Delete("user1", "file.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ = b.List()
	if len(list) != 0 {
		t.Fatalf("expected empty after delete")
	}
}

func TestMariaBackendScaffold(t *testing.T) {
	cfg := &config.Config{StorageDriver: "mariadb"}
	b, err := NewBackend(cfg, t.TempDir())
	if err != nil {
		t.Fatalf("new mariadb backend: %v", err)
	}
	if err := b.Save("user1", "dir/file.txt", bytes.NewReader([]byte("world"))); err != nil {
		t.Fatalf("save: %v", err)
	}
	f, err := b.Open("user1", "dir/file.txt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	buf := make([]byte, 5)
	if _, err := f.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	f.Close()
	if string(buf) != "world" {
		t.Fatalf("unexpected data %q", string(buf))
	}
	list, err := b.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("list mismatch: %#v %v", list, err)
	}
	if err := b.Delete("user1", "dir/file.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
