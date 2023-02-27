package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/client"
	"sykesdev.ca/yls/pkg/stream"
)

var secretsFile string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts an interactive instance of YLS",
	Long: `initiates an interactive instance of YLS where Streams are 
	configured and Youtube scheduled broadcasts will be deployed via the API`,
	Run: func(cmd *cobra.Command, args []string) {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM)
		signal.Notify(quit, syscall.SIGINT)
		ctx := context.Background()
		svc, err := initYoutubeService(ctx)
		if err != nil {
			YLSLogger().Fatal("unable to initialize the required Youtube Service endpoint", zap.String("error", err.Error()))
		}
		YLSLogger().Info("Youtube service has been initialized successfully")

		// get the streams from config
		var streams stream.StreamList
		if err := viper.Unmarshal(&streams); err != nil {
			YLSLogger().Fatal("viper was unable to unmarshal yaml config from file", zap.Error(err))
		}

		YLSLogger().Debug("streams", zap.Any("value", streams.Items))

		if len(streams.Items) == 0 {
			YLSLogger().Fatal("must specify at least one stream configuration to proceed")
		}

		c := cron.New()
		defer c.Stop()

		for _, s := range streams.Items {
			_, err := c.AddFunc(s.Schedule, s.WithService(svc).DryRun(dryRun).Go)
			if err != nil {
				YLSLogger().Fatal("failed to create scheduled job for Stream", zap.String("streamName", s.Name), zap.Error(err))
			}
			YLSLogger().Info("added new job to scheduler", zap.String("jobName", s.Name), zap.String("jobSchedule", s.Schedule))
		}

		YLSLogger().Info("starting cron scheduler")
		c.Start()

		sig := <-quit
		YLSLogger().Info("caught an exit signal. shutting down gracefully", zap.String("signal", sig.String()))
	},
}

func initYoutubeService(ctx context.Context) (*youtube.Service, error) {
	b, err := os.ReadFile(oauthConfigFile)
	if err != nil {
		YLSLogger().Fatal("Unable to read oauth configuration from file", zap.Error(err))
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		YLSLogger().Fatal("Unable to parse client secret file to config", zap.Error(err))
	}

	client := client.Get(ctx, secretsFile, config)
	svc, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func init() {
	startCmd.Flags().StringVar(&secretsFile, "secrets", "", "(required) The path to a JSON file containing OAuth2.0 Access and Refresh Tokens")
	startCmd.MarkFlagRequired("secrets")

	rootCmd.AddCommand(startCmd)
}
