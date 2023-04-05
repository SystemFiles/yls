package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sykesdev.ca/yls/pkg/logging"
)

// top-level CLI vars
var (
	loggingOut string
	dryRun     bool
	debugMode  bool
)

var (
	rootCmd = &cobra.Command{
		Use:   "yls",
		Short: "A tool that will periodically and automatically generate Youtube Livestream Broadcast Schedules",
	}
)

func init() {
	cobra.OnInitialize(initLogging)
	cobra.OnFinalize(cleanupLogging)
	cobra.EnableCaseInsensitive = true

	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "specifies whether YLS should be run in dry-run mode. This means YLS will make no changes, but will help evaluate changes that would be done")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "specifies whether Debug-level logs should be shown. This can be very noisy (be warned)")
	rootCmd.PersistentFlags().StringVarP(&loggingOut, "out", "o", os.Getenv("YLS_LOGGING_OUTPUT"), "specifies a file path to write logs to")
}

func initLogging() {
	if loggingOut != "" {
		YLSLogger(logging.LogPath{Value: loggingOut}).Info("logging to a file has been configured", zap.String("file", loggingOut))
	}
	if debugMode {
		YLSLogger(zap.DebugLevel).Debug("debug-mode has been enabled")
	}
}

func cleanupLogging() {
	YLSLogger().Sync()
}

func Execute() error {
	return rootCmd.Execute()
}
