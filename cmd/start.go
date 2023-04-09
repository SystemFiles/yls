package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/yaml.v3"
	"sykesdev.ca/yls/pkg/stream"
)

// youtube
var runNow bool
var streamConfigFile string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts an interactive instance of YLS",
	Run: func(cmd *cobra.Command, args []string) {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM)
		signal.Notify(quit, syscall.SIGINT)
		ctx := context.Background()

		streamUploader, err := stream.New(&stream.StreamUploaderConfig{
			Context:     ctx,
			OauthConfig: oauthConfigFile,
			Cache:       secretsCache,
			Scopes:      []string{youtube.YoutubeScope},
			DryRunMode:  dryRun,
		})
		if err != nil {
			YLSLogger().Fatal("failed to initialize Youtube Stream Uploader Client", zap.Error(err))
		}

		streams, err := getStreamsFromFile()
		if err != nil {
			YLSLogger().Fatal("unable to get streams from input file", zap.String("file", streamConfigFile), zap.Error(err))
		}

		if runNow {
			for _, s := range streams.Items {
				streamUploader.Upload(&s)() // this is kindof ugly :(
			}

			YLSLogger().Info("completed jobs for all configured streams", zap.Int("jobCount", len(streams.Items)))
			return
		}

		c := cron.New()
		for _, s := range streams.Items {
			_, err := c.AddFunc(s.Schedule, streamUploader.Upload(&s))
			if err != nil {
				YLSLogger().Fatal("failed to create scheduled job for Stream", zap.String("streamName", s.Name), zap.Error(err))
			}
			YLSLogger().Info("added new job to scheduler", zap.String("jobName", s.Name), zap.String("jobSchedule", s.Schedule))
		}

		YLSLogger().Info("starting scheduler")
		c.Start()

		sig := <-quit
		YLSLogger().Info("caught an exit signal. shutting down gracefully", zap.String("signal", sig.String()))
		stopCtx := c.Stop()
		<-stopCtx.Done()
	},
}

func getStreamsFromFile() (*stream.StreamList, error) {
	b, err := os.ReadFile(streamConfigFile)
	if err != nil {
		return nil, err
	}

	var streams stream.StreamList
	if err := yaml.Unmarshal(b, &streams); err != nil {
		return nil, err
	}
	YLSLogger().Debug("got streams from config file", zap.Any("value", streams.Items))

	if len(streams.Items) == 0 {
		return nil, errors.New("must specify at least one stream configuration to proceed")
	}

	return &streams, nil
}

func init() {
	startCmd.Flags().StringVarP(&streamConfigFile, "input", "i", "", "the path to the file which specifies configuration for youtube stream schedules")
	startCmd.Flags().BoolVarP(&runNow, "now", "n", false, "specifies whether to execute all configured stream jobs immediately instead of scheduling them for a future date/time. Note that any future jobs will NOT be scheduled when this flag is specified.")

	startCmd.MarkFlagRequired("input")
	rootCmd.AddCommand(startCmd)
}
