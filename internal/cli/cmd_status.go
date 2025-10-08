package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server health/status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		client := &http.Client{Timeout: 5 * time.Second}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.ServerURL+"/health", nil)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var m map[string]any
		json.NewDecoder(resp.Body).Decode(&m)
		fmt.Printf("health: %v (code=%d)\n", m, resp.StatusCode)
		// Try /status if token
		if c.AuthToken != "" {
			req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.ServerURL+"/status", nil)
			req2.Header.Set("Authorization", "Bearer "+c.AuthToken)
			resp2, err2 := client.Do(req2)
			if err2 == nil {
				defer resp2.Body.Close()
				var s map[string]any
				json.NewDecoder(resp2.Body).Decode(&s)
				fmt.Printf("status: %v (code=%d)\n", s, resp2.StatusCode)
			}
		}
		return nil
	},
}
