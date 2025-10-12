package cli

import (
	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/transfer"
	"git.tyss.io/cj3636/dman/pkg/model"
	"io"
)

// buildPublishTar creates a tar of changed (add/modify) files relative to user home.
func buildPublishTar(cfg *config.Config, changes []model.Change, w io.Writer) (int, error) {
	return transfer.BuildPublishTar(cfg, changes, w)
}

// applyInstallTar extracts files from an install tar stream into user home directories.
func applyInstallTar(cfg *config.Config, r io.Reader) (int, error) {
	return transfer.ApplyInstallTar(cfg, r)
}
