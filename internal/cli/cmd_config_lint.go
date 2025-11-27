package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var configLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate configuration and show effective tracking",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := requireConfig()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "configuration OK (storage_driver=%s)\n", cfg.StorageDriver)

		users := cfg.UsersList()
		sort.Slice(users, func(i, j int) bool { return users[i].Name < users[j].Name })
		for _, u := range users {
			includes, excludes := splitTrack(u.Track)
			fmt.Fprintf(cmd.OutOrStdout(), "- %s (%s): %d include(s), %d exclusion(s)\n", u.Name, u.Home, len(includes), len(excludes))
			if len(includes) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  includes: %s\n", strings.Join(includes, ", "))
			}
			if len(excludes) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  excludes: %s\n", strings.Join(excludes, ", "))
			}
		}
		return nil
	},
}

func splitTrack(patterns []string) (includes, excludes []string) {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "!") {
			excludes = append(excludes, strings.TrimPrefix(p, "!"))
			continue
		}
		includes = append(includes, p)
	}
	return includes, excludes
}

func init() { configCmd.AddCommand(configLintCmd) }
