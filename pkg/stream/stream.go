package stream

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/logging"
)

type StreamList struct {
	Items []Stream `mapstructure:"streams"`
}

type Stream struct {
	Name              string `mapstructure:"name"`
	TitlePrefix       string `mapstructure:"titlePrefix"`
	Description       string `mapstructure:"description"`
	Schedule          string `mapstructure:"schedule"`
	StartDelaySeconds uint16 `mapstructure:"delaySeconds"`
	PrivacyLevel      string `mapstructure:"privacyLevel"`

	service *youtube.Service `mapstructure:"-"`
	dryRun  bool             `mapstructure:"-"`
}

func (s *Stream) WithService(y *youtube.Service) *Stream {
	s.service = y
	return s
}

func (s *Stream) DryRun(val bool) *Stream {
	s.dryRun = val
	return s
}

func (s *Stream) Go() {
	title := fmt.Sprintf("%s (%s)", s.TitlePrefix, time.Now().Format(time.RFC3339))
	liveBroadcast := &youtube.LiveBroadcast{
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              title,
			Description:        s.Description,
			ScheduledStartTime: time.Now().Local().Add(time.Duration(s.StartDelaySeconds) * time.Second).Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus:           s.PrivacyLevel,
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

	if s.dryRun {
		logging.YLSLogger().Info("would have created LiveBroadcast resource, but is dry-run",
			zap.String("streamName", s.Name),
			zap.String("title", liveBroadcast.Snippet.Title),
			zap.String("description", liveBroadcast.Snippet.Description),
			zap.String("scheduledStart", liveBroadcast.Snippet.ScheduledStartTime),
			zap.String("privacyLevel", liveBroadcast.Status.PrivacyStatus),
		)
	} else {
		liveBroadcastCall := s.service.LiveBroadcasts.Insert([]string{"snippet", "status", "content_details"}, liveBroadcast)
		broadcastResp, err := liveBroadcastCall.Do()
		if err != nil {
			logging.YLSLogger().Error("failed to create a live broadcast", zap.String("streamName", s.Name), zap.Error(err))
			return
		}

		liveStreamCall := s.service.LiveStreams.List([]string{"cdn"}).Mine(true)
		existingLiveStreams, err := liveStreamCall.Do()
		if err != nil {
			logging.YLSLogger().Error("failed to get owned streams",
				zap.String("streamName", s.Name),
				zap.Error(err),
			)
		}

		streamKeys := []string{}
		for _, ls := range existingLiveStreams.Items {
			streamKeys = append(streamKeys, ls.Cdn.IngestionInfo.StreamName)
		}

		// Unfortunately it seems that the API cannot determine which Stream key will be used until a LS and Broadcast are bound.
		// Manually creating both can work, but will require a lot of extra detail in the configuration object which isn't ideal for users
		// For now just keep in mind: https://stackoverflow.com/questions/33798901/youtube-api-v3-get-live-now-rtmp-and-streamkey
		logging.YLSLogger().Info("created live scheduled broadcast",
			zap.String("streamName", s.Name),
			zap.String("broadcastName", broadcastResp.Snippet.Title),
			zap.String("scheduledStart", broadcastResp.Snippet.ScheduledStartTime),
			zap.String("currentStatus", broadcastResp.Status.RecordingStatus),
			zap.Strings("validStreamKeys", streamKeys),
		)
	}
}

func (s *Stream) String() string {
	return s.Name
}
