package server

import (
	"archive/tar"
	"io"
	"path/filepath"

	"git.tyss.io/cj3636/dman/internal/storage"
	"git.tyss.io/cj3636/dman/pkg/model"
)

// writeChangesTar writes add/modify/delete changes (filtered by provided predicate) into a tar stream.
// For install we pass predicate for ChangeDelete/ChangeModify.
func writeChangesTar(store storage.Backend, changes []model.Change, w io.Writer, include func(model.Change) bool) error {
	tw := tar.NewWriter(w)
	defer tw.Close()
	for _, ch := range changes {
		if !include(ch) {
			continue
		}
		f, err := store.Open(ch.User, ch.Path)
		if err != nil {
			continue
		}
		fi, _ := f.Stat()
		hdr, _ := tar.FileInfoHeader(fi, "")
		hdr.Name = ch.User + "/" + filepath.ToSlash(ch.Path)
		if err := tw.WriteHeader(hdr); err != nil {
			f.Close()
			continue
		}
		io.Copy(tw, f)
		f.Close()
	}
	return nil
}
