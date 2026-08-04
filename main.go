package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	faclient "github.com/sanderdescamps/go-purefa"
	"github.com/sanderdescamps/go-purefa-mock/mock"
	testdata "github.com/sanderdescamps/go-purefa-mock/mock/test-data"
	"github.com/spf13/cobra"
)

func main() {
	var (
		username    string
		apiToken    string
		dataDir     string
		testData    bool
		address     string
		port        int
		disableAuth bool
		endpoint    string
		apiversion  string
		insecure    bool
	)

	rootCmd := &cobra.Command{
		Use:   "go-purefa-mock",
		Short: "Mock Pure Storage FlashArray API",
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start the mock API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			var array *mock.Array
			var err error

			if dataDir != "" {
				array, err = mock.NewArrayFromDir(dataDir)
				if err != nil {
					return fmt.Errorf("failed to load test data: %w", err)
				}
			} else if testData {
				array, err = mock.NewArrayFromFS(testdata.TestData, ".")
				if err != nil {
					return fmt.Errorf("failed to load embedded test data: %w", err)
				}
			} else {
				array = mock.NewArray()
			}

			m := mock.NewMock(array)
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

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export from Pure storage array",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := faclient.NewRestClient(endpoint, apiToken, apiversion, "go-purefa-mock-export", "", false, insecure)
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			array, err := mock.NewArrayFromClient(client)
			if err != nil {
				return fmt.Errorf("failed to create mock array from client: %w", err)
			}

			if err := array.ExportToDir(dataDir); err != nil {
				return fmt.Errorf("failed to export array to directory: %w", err)
			}

			return nil
		},
	}

	exportCmd.Flags().StringVarP(&apiToken, "api-token", "t", "fake-auth-token", "API token for authentication")
	exportCmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "Directory path to store test data from")
	exportCmd.Flags().StringVarP(&endpoint, "endpoint", "n", "http://127.0.0.1:8080", "Pure Storage array endpoint")
	exportCmd.Flags().StringVarP(&apiversion, "version", "v", "latest", "API version to use")
	exportCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure server connections when using SSL")

	rootCmd.AddCommand(runCmd, exportCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
