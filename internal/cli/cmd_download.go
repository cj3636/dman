package cli

import (
	"context"
	"fmt"
	"git.tyss.io/cj3636/dman/internal/fsio"
	"git.tyss.io/cj3636/dman/internal/transfer"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <user> <path>",
	Short: "Download specific file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		user := args[0]
		rel := filepath.Clean(args[1])
		u, ok := c.Users[user]
		if !ok {
			return fmt.Errorf("unknown user: %s", user)
		}
		abs := filepath.Join(u.Home, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		client := transfer.New(c.ServerURL, c.AuthToken)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		rc, err := client.DownloadFile(ctx, user, filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		defer rc.Close()
		if err := fsio.AtomicWrite(abs, rc); err != nil {
			return err
		}
		fmt.Println("downloaded", user+":"+rel)
		return nil
	},
}
