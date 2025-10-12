package cli

import (
	"fmt"

	"git.tyss.io/cj3636/dman/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{Use: "version", Short: "Show build version", Run: func(cmd *cobra.Command, args []string) {
	fmt.Printf("dman version=%s commit=%s build_time=%s\n", buildinfo.Version, buildinfo.Commit, buildinfo.BuildTime)
}}

func init() { rootCmd.AddCommand(versionCmd) }
