package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Upload changed files to server",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		scanner := scan.New()
		inv, err := scanner.InventoryFor(c.UsersList())
		if err != nil {
			return err
		}
		reqBody := model.CompareRequest{Users: c.UserNames(), Inventory: inv}
		b, _ := json.Marshal(reqBody)
		httpClient := &http.Client{Timeout: 30 * time.Second}
		req, _ := http.NewRequest(http.MethodPost, c.ServerURL+"/compare", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if c.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.AuthToken)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var changes []model.Change
		if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
			return err
		}
		count := 0
		for _, ch := range changes {
			if ch.Type == model.ChangeAdd || ch.Type == model.ChangeModify { // client has file
				// find user spec
				u := c.Users[ch.User]
				abs := filepath.Join(u.Home, ch.Path)
				fi, err := os.Stat(abs)
				if err != nil || fi.IsDir() {
					continue
				}
				f, err := os.Open(abs)
				if err != nil {
					return err
				}
				upReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/upload?user=%s&path=%s", c.ServerURL, ch.User, filepath.ToSlash(ch.Path)), f)
				if c.AuthToken != "" {
					upReq.Header.Set("Authorization", "Bearer "+c.AuthToken)
				}
				ur, err := httpClient.Do(upReq)
				f.Close()
				if err != nil {
					return err
				}
				ur.Body.Close()
				if ur.StatusCode >= 300 {
					return fmt.Errorf("upload failed %s %s status=%d", ch.User, ch.Path, ur.StatusCode)
				}
				fmt.Printf("uploaded %s:%s\n", ch.User, ch.Path)
				count++
			}
		}
		fmt.Printf("publish complete (%d files)\n", count)
		return nil
	},
}
