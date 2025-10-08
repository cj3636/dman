package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
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
		client := &http.Client{}
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/upload?user=%s&path=%s", c.ServerURL, user, filepath.ToSlash(rel)), f)
		if c.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.AuthToken)
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("upload failed: %d", resp.StatusCode)
		}
		fmt.Println("uploaded", user+":"+rel)
		return nil
	},
}
