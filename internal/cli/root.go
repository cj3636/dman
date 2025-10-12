package cli

import (
	"fmt"
	"git.tyss.io/cj3636/dman/internal/config"
	"git.tyss.io/cj3636/dman/internal/logx"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	cfgPath  string
	cfg      *config.Config
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "dman",
	Short: "Dotfile sync tool (client & server)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cfg != nil { // already loaded (tests)
			return nil
		}
		if cmd.Name() == "init" { // allow missing config
			if _, err := os.Stat(cfgPath); err != nil { // missing: skip load
				return nil
			}
		}
		loaded, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		if err := loaded.Validate(); err != nil {
			return err
		}
		cfg = loaded
		return nil
	},
}

func Execute() error {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", defaultConfigPath(), "path to config file (dman.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (info|debug)")
	// register subcommands
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(serveCmd)
	// initCmd added via its own init()
	return rootCmd.Execute()
}

func mustLogger() *logx.Logger { return logx.New() }

func defaultConfigPath() string {
	// search current dir then home
	pwd, _ := os.Getwd()
	cand := filepath.Join(pwd, "dman.yaml")
	if _, err := os.Stat(cand); err == nil {
		return cand
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "dman.yaml")
}

func requireConfig() (*config.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	return cfg, nil
}
