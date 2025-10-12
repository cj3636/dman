package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"git.tyss.io/cj3636/dman/internal/transfer"
	"github.com/spf13/cobra"
)

var statusJSON bool

func init() { statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output JSON") }

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server health/status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		client := transfer.New(c.ServerURL, c.AuthToken)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h, err := client.Health(ctx)
		if err != nil {
			return err
		}
		if statusJSON {
			hb, _ := json.MarshalIndent(h, "", "  ")
			fmt.Println(string(hb))
		} else {
			fmt.Printf("health: %+v\n", h)
		}
		if c.AuthToken != "" {
			st, err := client.Status(ctx)
			if err == nil {
				if statusJSON {
					sb, _ := json.MarshalIndent(st, "", "  ")
					fmt.Println(string(sb))
				} else {
					fmt.Printf("status: %+v\n", st)
				}
			}
		}
		return nil
	},
}
