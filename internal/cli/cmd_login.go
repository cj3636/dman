package cli

import (
	"bufio"
	"fmt"
	"git.tyss.io/cj3636/dman/internal/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set auth token in config",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter token: ")
		tok, _ := reader.ReadString('\n')
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return fmt.Errorf("empty token")
		}
		c.AuthToken = tok
		return config.Save(c, cfgPath)
	},
}
