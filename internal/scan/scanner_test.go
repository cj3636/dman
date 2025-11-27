package scan

import (
	"git.tyss.io/cj3636/dman/pkg/model"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestInventoryFor(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".bashrc"), []byte("echo hi"), 0o644)
	s := New()
	inv, err := s.InventoryFor([]model.UserSpec{{Name: "u", Home: dir + "/", Track: []string{".bashrc"}}})
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

func TestInventoryForGlobAndExclusions(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, ".oh-my-zsh", "plugins", "git")
	os.MkdirAll(nested, 0o755)
	os.WriteFile(filepath.Join(nested, "git.plugin.zsh"), []byte("echo git"), 0o644)
	os.WriteFile(filepath.Join(nested, "README.md"), []byte("ignore me"), 0o644)

	docsDir := filepath.Join(dir, "docs")
	os.MkdirAll(docsDir, 0o755)
	os.WriteFile(filepath.Join(docsDir, "config.yaml"), []byte("secret"), 0o644)
	os.WriteFile(filepath.Join(docsDir, "guide.md"), []byte("keep"), 0o644)

	s := New()
	inv, err := s.InventoryFor([]model.UserSpec{{
		Name: "u", Home: dir + "/",
		Track: []string{".oh-my-zsh/plugins/**/*.zsh", "docs/", "!docs/config.yaml"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	if len(inv) != 2 {
		var paths []string
		for _, item := range inv {
			paths = append(paths, item.Path)
		}
		t.Fatalf("expected 2 tracked files got %d: %v", len(inv), paths)
	}
	paths := []string{inv[0].Path, inv[1].Path}
	sort.Strings(paths)

	if paths[0] != ".oh-my-zsh/plugins/git/git.plugin.zsh" {
		t.Fatalf("unexpected tracked file %s", paths[0])
	}
	if paths[1] != "docs/guide.md" {
		t.Fatalf("unexpected tracked file %s", paths[1])
	}
}

func TestInventorySupportsBraceExpansion(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "configs"), 0o755)
	os.WriteFile(filepath.Join(dir, "configs", "dev.yaml"), []byte("dev"), 0o644)
	os.WriteFile(filepath.Join(dir, "configs", "prod.yaml"), []byte("prod"), 0o644)

	s := New()
	inv, err := s.InventoryFor([]model.UserSpec{{
		Name: "u", Home: dir + "/",
		Track: []string{"configs/{dev,prod}.yaml"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(inv) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(inv))
	}
}
