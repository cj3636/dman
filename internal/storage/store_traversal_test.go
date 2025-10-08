package storage

import (
	"strings"
	"testing"
)

func TestStorePathTraversalRejected(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save("u", "../evil.txt", strings.NewReader("bad")); err == nil {
		t.Fatalf("expected traversal error")
	}
	if err := s.Save("u", "/abs.txt", strings.NewReader("bad")); err == nil {
		t.Fatalf("expected absolute path error")
	}
	if err := s.Save("u", "./ok.txt", strings.NewReader("ok")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	files, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range files {
		if f == "u/ok.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ok.txt stored, files=%v", files)
	}
}
