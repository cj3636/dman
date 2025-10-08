package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"git.tyss.io/cj3636/dman/internal/scan"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/spf13/cobra"
)

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
		b, _ := json.Marshal(reqBody)
		httpClient := &http.Client{Timeout: 15 * time.Second}
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
		for _, ch := range changes {
			fmt.Printf("%s\t%s\t%s\n", ch.Type, ch.User, ch.Path)
		}
		fmt.Printf("Total: %d changes\n", len(changes))
		return nil
	},
}
