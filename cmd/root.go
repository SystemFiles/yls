package cmd

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"sykesdev.ca/yls/pkg/logging"
)

// alias logger
var YLSLogger = logging.YLSLogger

// top-level CLI vars
var oauthConfigFile string
var streamConfigFile string
var dryRun bool
var debugMode bool

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

	rootCmd.PersistentFlags().StringVar(&oauthConfigFile, "oauth-config", "", "(required) the path to an associated OAuth configuration file (JSON) that is downloaded from Google for generation of the authorization token")
	rootCmd.PersistentFlags().StringVarP(&streamConfigFile, "stream-config", "c", "", "the path to the file which specifies configuration for youtube stream schedules")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "specifies whether YLS should be run in dry-run mode. This means YLS will make no changes, but will help evaluate changes that would be done")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "specifies whether Debug-level logs should be shown. This can be very noisy (be warned)")

	// required flags
	rootCmd.MarkPersistentFlagRequired("oauth-config")
}

func initConfig() {
	if streamConfigFile != "" {
		viper.SetConfigFile(streamConfigFile)
	} else {
		homeDir, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(homeDir)
		viper.SetConfigName(".yls")
		viper.SetConfigType("yaml")
		viper.SafeWriteConfig()
	}

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
