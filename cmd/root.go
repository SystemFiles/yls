package cmd

import (
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"sykesdev.ca/yls/pkg/logging"
)

// alias logger
var YLSLogger = logging.YLSLogger

// top-level CLI vars
var (
	authConfig   string
	secretsCache string
	headless     bool
	dryRun       bool
	debugMode    bool
)

// # root command
var (
	rootCmd = &cobra.Command{
		Use:   "yls",
		Short: "A tool that will periodically and automatically generate Youtube Livestream Broadcast Schedules",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnFinalize(cleanupLogging)
	cobra.EnableCaseInsensitive = true

	homeDir, err := os.UserHomeDir()
	if err != nil {
		YLSLogger().Fatal("unable to determine user home directory for config access")
	}

	rootCmd.PersistentFlags().StringVar(&secretsCache, "secrets-cache", path.Join(homeDir, ".youtube_oauth2_credentials"), "A path to a file location that will be used to cache OAuth2.0 Access and Refresh Tokens")
	rootCmd.PersistentFlags().StringVar(&authConfig, "auth-config", os.Getenv("YLS_AUTH_CONFIG"), "(required) the path to a Google OAuth2 configuration file (JSON) or a Private Key file (JSON) that contains information for authentication with Google APIs")
	rootCmd.PersistentFlags().BoolVar(&headless, "headless", os.Getenv("YLS_HEADLESS") != "" && os.Getenv("YLS_HEADLESS") != "false", "specifies whether YLS should use headless client protocols to authenticate a youtube service client")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "specifies whether YLS should be run in dry-run mode. This means YLS will make no changes, but will help evaluate changes that would be done")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "specifies whether Debug-level logs should be shown. This can be very noisy (be warned)")

	// required flags
	rootCmd.MarkPersistentFlagRequired("oauth-config")
}

func initConfig() {
	if debugMode {
		YLSLogger(zap.DebugLevel).Debug("debug-mode has been enabled") // init YLSLogger with debug logging/dev
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		YLSLogger().Debug("YLS configuration loaded")
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		YLSLogger().Info("YLS configuration changed", zap.String("event", e.Name))
	})
}

func cleanupLogging() {
	YLSLogger().Sync()
}

func Execute() error {
	return rootCmd.Execute()
}
