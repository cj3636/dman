package cli

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/internal/transfer"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/spf13/cobra"
)

var publishBulk bool
var publishPrune bool
var publishJSON bool
var publishGzip bool

func init() {
	publishCmd.Flags().BoolVar(&publishBulk, "bulk", false, "use tar bulk publish endpoint")
	publishCmd.Flags().BoolVar(&publishPrune, "prune", false, "delete server files missing locally")
	publishCmd.Flags().BoolVar(&publishJSON, "json", false, "output JSON summary")
	publishCmd.Flags().BoolVar(&publishGzip, "gzip", false, "gzip compress bulk tar payload")
}

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
		client := transfer.New(c.ServerURL, c.AuthToken)
		// overall publish timeout
		rootCtx, rootCancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer rootCancel()
		changes, err := client.Compare(rootCtx, reqBody, false)
		if err != nil {
			return err
		}
		if publishBulk {
			resultCh := make(chan struct {
				count int
				err   error
			}, 1)
			pr, pw := io.Pipe()
			enc := ""
			if publishGzip {
				enc = "gzip"
			}
			// build tar in background
			go func(enc string) {
				var w io.Writer = pw
				var gw *gzip.Writer
				if enc == "gzip" {
					gw = gzip.NewWriter(pw)
					w = gw
				}
				cnt, err := transfer.BuildPublishTar(c, changes, w)
				if gw != nil {
					gw.Close()
				}
				pw.CloseWithError(err)
				resultCh <- struct {
					count int
					err   error
				}{cnt, err}
			}(enc)
			pubErr := client.BulkPublish(rootCtx, pr, enc)
			res := <-resultCh
			if res.err != nil {
				return res.err
			}
			if pubErr != nil {
				return pubErr
			}
			if publishPrune {
				if _, err := client.Prune(rootCtx, changes); err != nil {
					return err
				}
			}
			if publishJSON {
				fmt.Printf("{\"files\":%d}\n", res.count)
			} else {
				fmt.Printf("bulk published %d files (stream)\n", res.count)
			}
			return nil
		}
		// non-bulk path: upload individually with per-file timeouts
		count := 0
		for _, ch := range changes {
			if ch.Type == model.ChangeAdd || ch.Type == model.ChangeModify {
				u := c.Users[ch.User]
				abs := filepath.Join(u.Home, ch.Path)
				fi, err := os.Stat(abs)
				if err != nil || fi.IsDir() { // skip missing/dir
					continue
				}
				f, err := os.Open(abs)
				if err != nil {
					return err
				}
				fileCtx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
				err = client.UploadFile(fileCtx, ch.User, filepath.ToSlash(ch.Path), f)
				cancel()
				f.Close()
				if err != nil {
					return err
				}
				if !publishJSON {
					fmt.Printf("uploaded %s:%s\n", ch.User, ch.Path)
				}
				count++
			}
		}
		if publishPrune {
			if _, err := client.Prune(rootCtx, changes); err != nil {
				return err
			}
		}
		if publishJSON {
			fmt.Printf("{\"files\":%d}\n", count)
		} else {
			fmt.Printf("publish complete (%d files)\n", count)
		}
		return nil
	},
}

// retained for backward compatibility with potential JSON output consumers (not used now)
// doPrune logic moved into transfer.Client.Prune
func legacyPrunePayload(changes []model.Change) []byte {
	var dels []map[string]string
	for _, ch := range changes {
		if ch.Type == model.ChangeDelete {
			dels = append(dels, map[string]string{"user": ch.User, "path": ch.Path})
		}
	}
	b, _ := json.Marshal(map[string]any{"deletes": dels})
	return b
}
