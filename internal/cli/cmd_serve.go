package cli

import (
	"context"
	"git.tyss.io/cj3636/dman/internal/logx"
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
		logger := logx.NewWithLevel(logLevel)
		srv, err := server.New(addr, c, logger)
		if err != nil {
			return err
		}
		go func() {
			logger.Info("server listening", "addr", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("server error", "err", err)
			}
		}()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logger.Info("server shutting down")
		return srv.Shutdown(ctx)
	},
}

func init() { serveCmd.Flags().StringVar(&serveAddr, "addr", ":7099", "listen address") }
