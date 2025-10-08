package scan

import (
	"git.tyss.io/cj3636/dman/pkg/model"
	"os"
	"path/filepath"
	"testing"
)

func TestInventoryFor(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".bashrc"), []byte("echo hi"), 0o644)
	s := New()
	inv, err := s.InventoryFor([]model.UserSpec{{Name: "u", Home: dir + "/", Include: []string{".bashrc"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(inv) != 1 {
		t.Fatalf("expected 1 got %d", len(inv))
	}
	if inv[0].Path != ".bashrc" {
		t.Fatalf("unexpected path %s", inv[0].Path)
	}
	if inv[0].Hash == "" {
		t.Fatalf("hash empty")
	}
}
