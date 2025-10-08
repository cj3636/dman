package cli

import (
	"fmt"
	"git.tyss.io/cj3636/dman/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear auth token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		c.AuthToken = ""
		fmt.Println("token cleared (remember to set again with 'dman login')")
		return config.Save(c, cfgPath)
	},
}
