package storage

import (
	"archive/tar"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestStoreBackupRestore(t *testing.T) {
	orig, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// seed files
	if err := orig.Save("alice", "configs/.bashrc", strings.NewReader("echo hi")); err != nil {
		t.Fatal(err)
	}
	if err := orig.Save("bob", "notes.txt", strings.NewReader("hello world")); err != nil {
		t.Fatal(err)
	}

	// backup
	var buf bytes.Buffer
	if err := orig.Backup(&buf); err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected backup data")
	}

	// quick sanity on tar stream (headers present)
	tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
	headers := 0
	for {
		_, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar read error: %v", err)
		}
		headers++
	}
	if headers != 2 {
		t.Fatalf("expected 2 headers got %d", headers)
	}

	// restore into new store
	restored, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := restored.Restore(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	files, err := restored.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files got %d", len(files))
	}

	// verify content of one file
	f, err := restored.Open("alice", "configs/.bashrc")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(f)
	f.Close()
	if string(b) != "echo hi" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}
