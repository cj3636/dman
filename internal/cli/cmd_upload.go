package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git.tyss.io/cj3636/dman/internal/transfer"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <user> <path>",
	Short: "Upload specific file",
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
		f, err := os.Open(abs)
		if err != nil {
			return err
		}
		defer f.Close()
		client := transfer.New(c.ServerURL, c.AuthToken)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.UploadFile(ctx, user, filepath.ToSlash(rel), f); err != nil {
			return err
		}
		fmt.Println("uploaded", user+":"+rel)
		return nil
	},
}
