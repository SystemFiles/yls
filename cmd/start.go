package cmd

import (
	"context"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"sykesdev.ca/yls/pkg/stream"
)

// youtube
var runNow bool
var streamConfigFile string

// publish
var publish bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts an interactive instance of YLS",
	Run: func(cmd *cobra.Command, args []string) {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM)
		signal.Notify(quit, syscall.SIGINT)
		ctx := context.Background()
		scheduler, err := getSchedulerConfigFromFile(ctx)
		if err != nil {
			YLSLogger().Fatal("unable to get streams from input file", zap.String("file", streamConfigFile), zap.Error(err))
		}

		if runNow {
			jobs := 0
			for scheduler.Streams.HasNext() {
				stream.ScheduleLiveStream(&stream.ScheduleOptions{
					Publisher: scheduler.Publisher,
					Service:   scheduler.Svc,
					Stream:    scheduler.Streams.Next(),
					DryRun:    dryRun,
					Publish:   publish,
				})
				jobs++
			}

			YLSLogger().Info("completed jobs for all configured streams", zap.Int("jobCount", jobs))
			return
		}

		c := cron.New()
		for scheduler.Streams.HasNext() {
			s := scheduler.Streams.Next()
			_, err := c.AddFunc(s.Schedule, stream.ScheduleLiveStream(&stream.ScheduleOptions{
				Publisher: scheduler.Publisher,
				Service:   scheduler.Svc,
				Stream:    scheduler.Streams.Next(),
				DryRun:    dryRun,
				Publish:   publish,
			}))
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

func getSchedulerConfigFromFile(ctx context.Context) (*stream.StreamScheduler, error) {
	b, err := os.ReadFile(streamConfigFile)
	if err != nil {
		return nil, err
	}

	var config stream.StreamSchedulerConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return stream.NewScheduler(ctx, &config)
}

func init() {
	h, _ := os.UserHomeDir()

	startCmd.Flags().StringVarP(&streamConfigFile, "input", "i", path.Join(h, ".yls.yaml"), "the path to the file which specifies configuration for youtube stream schedules (default '$HOME/.yls.yaml')")
	startCmd.Flags().BoolVarP(&runNow, "now", "n", false, "specifies whether to execute all configured stream jobs immediately instead of scheduling them for a future date/time. Note that any future jobs will NOT be scheduled when this flag is specified.")

	rootCmd.AddCommand(startCmd)
}
