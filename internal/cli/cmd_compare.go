package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/internal/transfer"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/spf13/cobra"
)

var (
	compareShowSame bool
	compareJSON     bool
)

func init() {
	compareCmd.Flags().BoolVar(&compareShowSame, "show-same", false, "include unchanged entries")
	compareCmd.Flags().BoolVar(&compareJSON, "json", false, "output JSON")
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare local vs server",
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		changes, err := client.Compare(ctx, reqBody, compareShowSame)
		if err != nil {
			return err
		}
		if compareJSON {
			out, _ := json.MarshalIndent(changes, "", "  ")
			fmt.Println(string(out))
			return nil
		}
		for _, ch := range changes {
			fmt.Printf("%s\t%s\t%s\n", ch.Type, ch.User, ch.Path)
		}
		fmt.Printf("Total: %d changes\n", len(changes))
		return nil
	},
}
