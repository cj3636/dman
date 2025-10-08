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

var publishBulk bool

func init() { publishCmd.Flags().BoolVar(&publishBulk, "bulk", false, "use tar bulk publish endpoint") }

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
		bReq, _ := json.Marshal(reqBody)
		httpClient := &http.Client{Timeout: 60 * time.Second}
		if publishBulk {
			// first get change set
			cmpReq, _ := http.NewRequest(http.MethodPost, c.ServerURL+"/compare", bytes.NewReader(bReq))
			cmpReq.Header.Set("Content-Type", "application/json")
			if c.AuthToken != "" {
				cmpReq.Header.Set("Authorization", "Bearer "+c.AuthToken)
			}
			cmpResp, err := httpClient.Do(cmpReq)
			if err != nil {
				return err
			}
			defer cmpResp.Body.Close()
			var changes []model.Change
			if err := json.NewDecoder(cmpResp.Body).Decode(&changes); err != nil {
				return err
			}
			var tarBuf bytes.Buffer
			count, err := buildPublishTar(c, changes, &tarBuf)
			if err != nil {
				return err
			}
			pubReq, _ := http.NewRequest(http.MethodPost, c.ServerURL+"/publish", bytes.NewReader(tarBuf.Bytes()))
			pubReq.Header.Set("Content-Type", "application/x-tar")
			if c.AuthToken != "" {
				pubReq.Header.Set("Authorization", "Bearer "+c.AuthToken)
			}
			pubResp, err := httpClient.Do(pubReq)
			if err != nil {
				return err
			}
			defer pubResp.Body.Close()
			if pubResp.StatusCode >= 300 {
				return fmt.Errorf("publish failed status=%d", pubResp.StatusCode)
			}
			fmt.Printf("bulk published %d files\n", count)
			return nil
		}
		// non-bulk path (legacy per-file uploads via compare)
		cmpReq, _ := http.NewRequest(http.MethodPost, c.ServerURL+"/compare", bytes.NewReader(bReq))
		cmpReq.Header.Set("Content-Type", "application/json")
		if c.AuthToken != "" {
			cmpReq.Header.Set("Authorization", "Bearer "+c.AuthToken)
		}
		cmpResp, err := httpClient.Do(cmpReq)
		if err != nil {
			return err
		}
		defer cmpResp.Body.Close()
		var changes []model.Change
		if err := json.NewDecoder(cmpResp.Body).Decode(&changes); err != nil {
			return err
		}
		count := 0
		for _, ch := range changes {
			if ch.Type == model.ChangeAdd || ch.Type == model.ChangeModify {
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
					return fmt.Errorf("upload failed %s status=%d", ch.Path, ur.StatusCode)
				}
				fmt.Printf("uploaded %s:%s\n", ch.User, ch.Path)
				count++
			}
		}
		fmt.Printf("publish complete (%d files)\n", count)
		return nil
	},
}
