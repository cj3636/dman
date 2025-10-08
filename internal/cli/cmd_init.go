package cli

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"strings"

	"git.tyss.io/cj3636/dman/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive first-time configuration setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		// decide path
		path := cfgPath
		if path == "" {
			path = defaultConfigPath()
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Server URL [http://localhost:7099]: ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)
		if url == "" {
			url = "http://localhost:7099"
		}
		fmt.Printf("Auth token (leave blank to auto-generate): ")
		tok, _ := reader.ReadString('\n')
		tok = strings.TrimSpace(tok)
		if tok == "" {
			tok = randomToken(24)
		}
		u, _ := user.Current()
		home := ""
		if u != nil {
			home = u.HomeDir
		}
		// minimal user include set
		cfgObj := &config.Config{AuthToken: tok, ServerURL: url, Users: map[string]config.User{}}
		if home != "" {
			cfgObj.Users[u.Username] = config.User{Home: ensureTrailingSlash(home), Include: []string{".bashrc", ".zshrc"}}
		}
		if err := config.Save(cfgObj, path); err != nil {
			return err
		}
		fmt.Println("Config written to", path)
		return nil
	},
}

func init() { rootCmd.AddCommand(initCmd) }

func randomToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ensureTrailingSlash(s string) string {
	if s == "" {
		return s
	}
	if strings.HasSuffix(s, string(os.PathSeparator)) {
		return s
	}
	return s + string(os.PathSeparator)
}
