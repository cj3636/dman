package cli

import (
	"context"
	"fmt"
	"git.tyss.io/cj3636/dman/internal/server"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var serveAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireConfig()
		if err != nil {
			return err
		}
		addr := serveAddr
		if addr == "" {
			addr = ":7099"
		}
		srv, err := server.New(addr, c)
		if err != nil {
			return err
		}
		go func() {
			fmt.Printf("server listening on %s\n", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println("server error:", err)
			}
		}()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}

func init() { serveCmd.Flags().StringVar(&serveAddr, "addr", ":7099", "listen address") }
