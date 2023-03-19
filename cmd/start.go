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
	"sykesdev.ca/yls/pkg/pub"
	"sykesdev.ca/yls/pkg/stream"
)

// youtube
var runNow bool
var streamConfigFile string

// publish
var publish bool
var publishOptions = &pub.PublishOptions{}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts an interactive instance of YLS",
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

		if runNow {
			for _, s := range streams.Items {
				if publish {
					s.WithService(svc).DryRun(dryRun).WithPublisher(publishOptions).Go()
				} else {
					s.WithService(svc).DryRun(dryRun).Go()
				}
			}

			YLSLogger().Info("completed jobs for all configured streams")
			return
		}

		for _, s := range streams.Items {
			var err error
			if publish {
				_, err = c.AddFunc(s.Schedule, s.WithService(svc).WithPublisher(publishOptions).DryRun(dryRun).Go)
			} else {
				_, err = c.AddFunc(s.Schedule, s.WithService(svc).DryRun(dryRun).Go)
			}
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

	client := client.Get(ctx, secretsCache, config)
	svc, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func init() {
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

	startCmd.Flags().StringVarP(&streamConfigFile, "input", "i", "", "the path to the file which specifies configuration for youtube stream schedules (default '$HOME/.yls.yaml')")
	startCmd.Flags().BoolVarP(&runNow, "now", "n", false, "specifies whether to execute all configured stream jobs immediately instead of scheduling them for a future date/time. Note that any future jobs will NOT be scheduled when this flag is specified.")
	startCmd.Flags().BoolVarP(&publish, "publish", "p", false, "specifies whether to publish the stream using a configured publisher (ie: wordpress)")
	startCmd.Flags().StringVar(&publishOptions.WPConfig, "wp-config", os.Getenv("YLS_WP_CONFIG"), "(optional) the path to a file containing configuration (in YAML) for a wordpress publisher")
	startCmd.Flags().StringVar(&publishOptions.WPBaseURL, "wp-base-url", os.Getenv("YLS_WP_BASE_URL"), "the base URL for a the wordpress v2 API")
	startCmd.Flags().StringVar(&publishOptions.WPUsername, "wp-username", os.Getenv("YLS_WP_USERNAME"), "the username for the user or service account to use for wordpress publishing")
	startCmd.Flags().StringVar(&publishOptions.WPAppToken, "wp-app-token", os.Getenv("YLS_WP_APP_TOKEN"), "the wordpress App token to use to authenticate the identified wordpress user")
	startCmd.Flags().IntVar(&publishOptions.WPExistingPageId, "wp-page-id", 0, "(optional) a page ID for a wordpress page to publish changes to (if not specified, a page will be created)")
	startCmd.Flags().StringVar(&publishOptions.WPPageTemplate, "wp-page-template", os.Getenv("YLS_WP_PAGE_TEMPLATE"), "a string that contains a gotemplate-compatible HTLM page template to use to construct wordpress page content")

	rootCmd.AddCommand(startCmd)
}
