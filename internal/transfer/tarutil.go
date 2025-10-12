package transfer

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"

	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/pkg/model"
)

// BuildPublishTar writes add/modify files to w and returns count.
func BuildPublishTar(cfg *config.Config, changes []model.Change, w io.Writer) (int, error) {
	tw := tar.NewWriter(w)
	count := 0
	for _, ch := range changes {
		if ch.Type != model.ChangeAdd && ch.Type != model.ChangeModify {
			continue
		}
		u, ok := cfg.Users[ch.User]
		if !ok {
			continue
		}
		abs := filepath.Join(u.Home, ch.Path)
		fi, err := os.Stat(abs)
		if err != nil || fi.IsDir() {
			continue
		}
		f, err := os.Open(abs)
		if err != nil {
			continue
		}
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			f.Close()
			continue
		}
		hdr.Name = ch.User + "/" + filepath.ToSlash(ch.Path)
		if err := tw.WriteHeader(hdr); err != nil {
			f.Close()
			continue
		}
		io.Copy(tw, f)
		f.Close()
		count++
	}
	if err := tw.Close(); err != nil {
		return count, err
	}
	return count, nil
}

// ApplyInstallTar extracts tar entries (user/relpath) into user homes.
func ApplyInstallTar(cfg *config.Config, r io.Reader) (int, error) {
	tr := tar.NewReader(r)
	written := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return written, err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		parts := strings.SplitN(filepath.ToSlash(hdr.Name), "/", 2)
		if len(parts) != 2 {
			continue
		}
		u, ok := cfg.Users[parts[0]]
		if !ok {
			continue
		}
		abs := filepath.Join(u.Home, parts[1])
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return written, err
		}
		f, err := os.Create(abs)
		if err != nil {
			return written, err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return written, err
		}
		f.Close()
		written++
	}
	return written, nil
}
