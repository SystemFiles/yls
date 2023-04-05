package stream

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/logging"
	"sykesdev.ca/yls/pkg/pub"
)

type StreamList struct {
	Streams []Stream `yaml:"streams"`
}

type Stream struct {
	Name              string `yaml:"name"`
	Title             string `yaml:"title"`
	Description       string `yaml:"description"`
	Schedule          string `yaml:"schedule"`
	StartDelaySeconds uint16 `yaml:"delaySeconds"`
	PrivacyLevel      string `yaml:"privacyLevel"`
}

type ScheduleOptions struct {
	pub.Publisher
	Service *youtube.Service
	Stream  *Stream
	DryRun  bool
	Publish bool
}

func ScheduleLiveStream(opts *ScheduleOptions) func() {
	return func() {
		liveBroadcast := &youtube.LiveBroadcast{
			Snippet: &youtube.LiveBroadcastSnippet{
				Title:              opts.Stream.Title,
				Description:        opts.Stream.Description,
				ScheduledStartTime: time.Now().Local().Add(time.Duration(opts.Stream.StartDelaySeconds) * time.Second).Format(time.RFC3339),
			},
			Status: &youtube.LiveBroadcastStatus{
				PrivacyStatus:           opts.Stream.PrivacyLevel,
				MadeForKids:             false,
				SelfDeclaredMadeForKids: false,
			},
			ContentDetails: &youtube.LiveBroadcastContentDetails{
				EnableClosedCaptions: false,
				EnableAutoStop:       true,
				EnableDvr:            true,
				ClosedCaptionsType:   "closedCaptionsDisabled",
			},
		}

		if opts.DryRun {
			logging.YLSLogger().Info("would have created LiveBroadcast resource, but is dry-run",
				zap.String("streamName", opts.Stream.Name),
				zap.String("title", liveBroadcast.Snippet.Title),
				zap.String("description", liveBroadcast.Snippet.Description),
				zap.String("scheduledStart", liveBroadcast.Snippet.ScheduledStartTime),
				zap.String("privacyLevel", liveBroadcast.Status.PrivacyStatus),
			)
		} else {
			liveBroadcastCall := opts.Service.LiveBroadcasts.Insert([]string{"snippet", "status", "content_details"}, liveBroadcast)
			broadcastResp, err := liveBroadcastCall.Do()
			if err != nil {
				logging.YLSLogger().Error("failed to create a live broadcast", zap.String("streamName", opts.Stream.Name), zap.Error(err))
				return
			}

			liveStreamCall := opts.Service.LiveStreams.List([]string{"cdn"}).Mine(true)
			existingLiveStreams, err := liveStreamCall.Do()
			if err != nil {
				logging.YLSLogger().Error("failed to get owned streams",
					zap.String("streamName", opts.Stream.Name),
					zap.Error(err),
				)
				return
			}

			streamKeys := []string{}
			for _, ls := range existingLiveStreams.Items {
				streamKeys = append(streamKeys, ls.Cdn.IngestionInfo.StreamName)
			}

			// Unfortunately it seems that the API cannot determine which Stream key will be used until a LS and Broadcast are bound.
			// Manually creating both can work, but will require a lot of extra detail in the configuration object which isn't ideal for users
			// For now just keep in mind: https://stackoverflow.com/questions/33798901/youtube-api-v3-get-live-now-rtmp-and-streamkey
			logging.YLSLogger().Info("created live scheduled broadcast",
				zap.String("streamName", opts.Stream.Name),
				zap.String("broadcastName", broadcastResp.Snippet.Title),
				zap.String("scheduledStart", broadcastResp.Snippet.ScheduledStartTime),
				zap.String("currentStatus", broadcastResp.Status.RecordingStatus),
				zap.Strings("validStreamKeys", streamKeys),
				zap.String("shareableLink", fmt.Sprintf("https://youtube.com/live/%s?feature=share", broadcastResp.Id)),
				zap.String("embedableLink", fmt.Sprintf("https://youtube.com/embed/%s", broadcastResp.Id)),
			)

			if opts.Publish {
				logging.YLSLogger().Info("publishing stream details to using configured publisher",
					zap.String("streamName", opts.Stream.Name),
				)

				if err := opts.Publisher.Publish(liveBroadcast, opts.Publisher); err != nil {
					logging.YLSLogger().Error("failed to publish stream details using configured publisher",
						zap.String("streamName", opts.Stream.Name),
						zap.Error(err),
					)
				}
			}
		}
	}
}

type Collection interface {
	iterator() Iterator
}

type Iterator interface {
	HasNext() bool
	Next() *Stream
}

type StreamIterator struct {
	index   int
	streams []*Stream
}

func (s *StreamIterator) HasNext() bool {
	return s.index < len(s.streams)
}

func (s *StreamIterator) Next() *Stream {
	if s.HasNext() {
		s.index++
		return s.streams[s.index]
	}
	return nil
}
