package main

import (
	"fmt"
	"log/slog"

	faclient "github.com/sanderdescamps/go-purefa"
	"github.com/sanderdescamps/go-purefa-mock/internal/fakearray"
	"github.com/spf13/cobra"
)

type Command interface {
	RegisterFlags(cmd *cobra.Command)
}

type ExportCommand struct {
	*cobra.Command
}

func NewExportCommand() *cobra.Command {
	var (
		apiToken   string
		exportDir  string
		endpoint   string
		apiversion string
		insecure   bool
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data from the mock API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := slog.New(NewCobraLogHandler(cmd.OutOrStdout(), getLogLevel()))
			logger.Info("Set logger", "level", getLogLevel())
			cfg := faclient.ClientConfig{
				ApiVersion: apiversion,
				UserAgent:  "go-purefa-mock-export",
				Insecure:   insecure,
				RequestID:  "",
				Debug:      false,
			}

			client, err := faclient.NewRestClient(endpoint, apiToken, cfg)
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			if err := fakearray.ExportArrayToDir(client, exportDir); err != nil {
				return fmt.Errorf("failed to export array to directory: %w", err)
			}

			return nil
		},
	}

	c := cmd
	c.Flags().StringVarP(&apiToken, "api-token", "t", "fake-auth-token", "API token for authentication")
	c.Flags().StringVarP(&exportDir, "export-dir", "d", "", "Directory path to store exported data")
	c.Flags().StringVarP(&endpoint, "endpoint", "n", "http://127.0.0.1:8080", "Pure Storage array endpoint")
	c.Flags().StringVarP(&apiversion, "version", "v", "latest", "API version to use")
	c.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure server connections when using SSL")

	return cmd
}
