package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"git.tyss.io/cj3636/dman/internal/fsio"
	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/spf13/cobra"
)

var installBulk bool

func init() { installCmd.Flags().BoolVar(&installBulk, "bulk", false, "use tar bulk install endpoint") }

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download newer server files",
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
		if installBulk {
			instReq, _ := http.NewRequest(http.MethodPost, c.ServerURL+"/install", bytes.NewReader(bReq))
			instReq.Header.Set("Content-Type", "application/json")
			if c.AuthToken != "" {
				instReq.Header.Set("Authorization", "Bearer "+c.AuthToken)
			}
			resp, err := httpClient.Do(instReq)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				return fmt.Errorf("install failed status=%d", resp.StatusCode)
			}
			count, err := applyInstallTar(c, resp.Body)
			if err != nil {
				return err
			}
			fmt.Printf("bulk installed %d files\n", count)
			return nil
		}
		// legacy per-file path
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
			if ch.Type == model.ChangeDelete || ch.Type == model.ChangeModify {
				u := c.Users[ch.User]
				abs := filepath.Join(u.Home, ch.Path)
				if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
					return err
				}
				dlReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/download?user=%s&path=%s", c.ServerURL, ch.User, filepath.ToSlash(ch.Path)), nil)
				if c.AuthToken != "" {
					dlReq.Header.Set("Authorization", "Bearer "+c.AuthToken)
				}
				dlResp, err := httpClient.Do(dlReq)
				if err != nil {
					return err
				}
				if dlResp.StatusCode == 404 {
					dlResp.Body.Close()
					continue
				}
				if dlResp.StatusCode >= 300 {
					dlResp.Body.Close()
					return fmt.Errorf("download failed %s status=%d", ch.Path, dlResp.StatusCode)
				}
				if err := fsio.AtomicWrite(abs, dlResp.Body); err != nil {
					dlResp.Body.Close()
					return err
				}
				dlResp.Body.Close()
				fmt.Printf("downloaded %s:%s\n", ch.User, ch.Path)
				count++
			}
		}
		fmt.Printf("install complete (%d files)\n", count)
		return nil
	},
}
