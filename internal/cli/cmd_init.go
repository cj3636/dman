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
		path := cfgPath
		if path == "" {
			path = defaultConfigPath()
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}
		reader := bufio.NewReader(os.Stdin)

		// Hostname & Port
		fmt.Printf("Hostname [localhost]: ")
		host, _ := reader.ReadString('\n')
		host = strings.TrimSpace(host)
		if host == "" {
			host = "localhost"
		}
		fmt.Printf("Port [7099]: ")
		port, _ := reader.ReadString('\n')
		port = strings.TrimSpace(port)
		if port == "" {
			port = "7099"
		}
		url := fmt.Sprintf("http://%s:%s", host, port)

		// Auth token
		fmt.Printf("Auth token (leave blank to auto-generate): ")
		tok, _ := reader.ReadString('\n')
		tok = strings.TrimSpace(tok)
		if tok == "" {
			tok = randomToken(24)
		}

		// Users
		fmt.Printf("Users (comma separated) [root,cjserver]: ")
		userLine, _ := reader.ReadString('\n')
		userLine = strings.TrimSpace(userLine)
		if userLine == "" {
			userLine = "root,cjserver"
		}
		userParts := splitCommaList(userLine)
		usersMap := map[string]config.User{}
		for _, un := range userParts {
			if un == "root" {
				usersMap[un] = config.User{Home: "/root/"}
			} else {
				usersMap[un] = config.User{Home: "/home/" + un + "/"}
			}
		}

		// Includes
		fmt.Println("Default include list (comma separated) - press Enter to accept defaults:")
		fmt.Println(strings.Join(config.DefaultInclude, ","))
		fmt.Printf("Include overrides: ")
		incLine, _ := reader.ReadString('\n')
		incLine = strings.TrimSpace(incLine)
		var include []string
		if incLine == "" {
			include = config.DefaultInclude
		} else {
			include = splitCommaList(incLine)
		}

		cfgObj := &config.Config{AuthToken: tok, ServerURL: url, GlobalInclude: include, Users: usersMap}
		if err := config.Save(cfgObj, path); err != nil {
			return err
		}
		fmt.Println("Config written to", path)
		fmt.Println("You can edit per-user include overrides later by adding an 'include:' list under a user.")
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

func splitCommaList(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
