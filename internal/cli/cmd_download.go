package cli

import (
	"fmt"
	"git.tyss.io/cj3636/dman/internal/fsio"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
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
		client := &http.Client{}
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/download?user=%s&path=%s", c.ServerURL, user, filepath.ToSlash(rel)), nil)
		if c.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.AuthToken)
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("download failed: %d", resp.StatusCode)
		}
		if err := fsio.AtomicWrite(abs, resp.Body); err != nil {
			return err
		}
		fmt.Println("downloaded", user+":"+rel)
		return nil
	},
}
