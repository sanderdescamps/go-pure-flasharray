package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/mock"
	"github.com/spf13/cobra"
)

func NewRunCommand() *cobra.Command {
	var (
		username    string
		apiToken    string
		dataDir     string
		testData    bool
		address     string
		port        int
		disableAuth bool
	)

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start the mock API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := slog.New(NewCobraLogHandler(cmd.OutOrStdout(), getLogLevel()))
			logger.Info("Set logger", "level", getLogLevel())

			var a *fakearray.Array
			var err error

			if dataDir != "" {
				a, err = fakearray.NewArrayFromDir(dataDir)
				if err != nil {
					return fmt.Errorf("failed to load test data: %w", err)
				}
			} else if testData {
				a, err = fakearray.NewArrayWithTestData(fakearray.WithLogger(logger))
				if err != nil {
					return fmt.Errorf("failed to load embedded test data: %w", err)
				}
			} else {
				a = fakearray.NewArray(fakearray.WithLogger(logger))
			}

			m := mock.NewMock(a, mock.WithLogger(logger))
			m.Username = username
			m.APIToken = apiToken
			m.DisableAuth = disableAuth

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			serverErr := make(chan error, 1)
			go func() {
				serverErr <- m.Start(address, port)
			}()

			select {
			case sig := <-quit:
				fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := m.Shutdown(ctx); err != nil {
					return fmt.Errorf("shutdown error: %w", err)
				}
				return nil
			case err := <-serverErr:
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}
				return nil
			}
		},
	}

	runCmd.Flags().StringVarP(&username, "username", "u", "pureuser", "Username for API authentication")
	runCmd.Flags().StringVarP(&apiToken, "api-token", "t", "fake-auth-token", "API token for authentication")
	runCmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "Directory path to load test data from")
	runCmd.Flags().BoolVarP(&testData, "test-data", "e", false, "Use embedded test data")
	runCmd.Flags().StringVarP(&address, "address", "a", "127.0.0.1", "Server listen address")
	runCmd.Flags().IntVarP(&port, "port", "p", 8080, "Server listen port")
	runCmd.Flags().BoolVarP(&disableAuth, "disable-auth", "s", false, "Disable authentication")
	runCmd.MarkFlagsMutuallyExclusive("data-dir", "test-data")

	return runCmd
}
