package cli

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"git.tyss.io/cj3636/dman/internal/fsio"
	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/internal/transfer"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/spf13/cobra"
)

var installBulk bool
var installJSON bool
var installGzip bool

func init() {
	installCmd.Flags().BoolVar(&installBulk, "bulk", false, "use tar bulk install endpoint")
	installCmd.Flags().BoolVar(&installJSON, "json", false, "output JSON summary")
	installCmd.Flags().BoolVar(&installGzip, "gzip", false, "request gzip compressed bulk tar")
}

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
		client := transfer.New(c.ServerURL, c.AuthToken)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if installBulk {
			bodyReq := reqBody
			respBody, err := client.BulkInstall(ctx, bodyReq, func() string {
				if installGzip {
					return "gzip"
				}
				return ""
			}())
			if err != nil {
				return err
			}
			defer respBody.Close()
			var reader io.Reader = respBody
			if installGzip {
				if gr, err := gzip.NewReader(respBody); err == nil {
					reader = gr
					defer gr.Close()
				}
			}
			count, err := applyInstallTar(c, reader)
			if err != nil {
				return err
			}
			if installJSON {
				fmt.Printf("{\"files\":%d}\n", count)
			} else {
				fmt.Printf("bulk installed %d files\n", count)
			}
			return nil
		}
		changes, err := client.Compare(ctx, reqBody, false)
		if err != nil {
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
				fileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				rc, err := client.DownloadFile(fileCtx, ch.User, filepath.ToSlash(ch.Path))
				if err != nil {
					cancel()
					return err
				}
				if err := fsio.AtomicWrite(abs, rc); err != nil {
					rc.Close()
					cancel()
					return err
				}
				rc.Close()
				cancel()
				if !installJSON {
					fmt.Printf("downloaded %s:%s\n", ch.User, ch.Path)
				}
				count++
			}
		}
		if installJSON {
			fmt.Printf("{\"files\":%d}\n", count)
		} else {
			fmt.Printf("install complete (%d files)\n", count)
		}
		return nil
	},
}
