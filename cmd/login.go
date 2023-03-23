package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/client"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "forces a login and stores the resulting access token in the configured credentials cache",
	Long:  "by forcing a login, we can update credentials before use in a docker/container environment for example",
	Run: func(cmd *cobra.Command, args []string) {
		if oauthConfigFile == "" {
			YLSLogger().Fatal("oauth configuration file is required. specify --oauth-config or use the environment variable YLS_OAUTH_CONFIG")
		}
		b, err := os.ReadFile(oauthConfigFile)
		if err != nil {
			YLSLogger().Fatal("unable to read oauth configuration from file", zap.Error(err))
		}

		config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
		if err != nil {
			YLSLogger().Fatal("unable to parse client secret file to config", zap.Error(err))
		}
		c := client.Get(context.TODO(), secretsCache, config)
		if c == nil {
			YLSLogger().Fatal("login failed")
		}

		YLSLogger().Info("login succeeded!")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
