package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var logLevelStr string = "info"

func getLogLevel() slog.Level {
	switch strings.ToLower(logLevelStr) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "go-purefa-mock",
		Short: "Mock Pure Storage FlashArray API",
	}

	runCmd := NewRunCommand()
	exportCmd := NewExportCommand()

	rootCmd.PersistentFlags().StringVarP(&logLevelStr, "log", "l", "info", "Set the log level (debug, info, warn, error)")

	rootCmd.AddCommand(runCmd, exportCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func main() {
	Execute()
}
