package stream

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/logging"
	"sykesdev.ca/yls/pkg/pub"
)

type UploaderPublisher interface {
	Upload(*Stream) func()
}

type Builder interface {
	SetDryRun()
	SetService(*youtube.Service)
	Build() *StreamUploadClient
}

type StreamUploadClientBuilder struct {
	service *youtube.Service
	dryRun  bool
}

func NewStreamUploadClientBuilder() *StreamUploadClientBuilder {
	return &StreamUploadClientBuilder{}
}

func (b *StreamUploadClientBuilder) SetDryRun() {
	b.dryRun = true
}

func (b *StreamUploadClientBuilder) SetService(y *youtube.Service) {
	b.service = y
}

func (b *StreamUploadClientBuilder) Build() *StreamUploadClient {
	return &StreamUploadClient{
		svc:    b.service,
		dryRun: b.dryRun,
	}
}

type StreamUploadClient struct {
	svc    *youtube.Service
	dryRun bool
}

func (u *StreamUploadClient) Upload(s *Stream) func() {
	return func() {
		if u.svc == nil {
			logging.YLSLogger().Error("unable to create Live Broadcast resource. no service was available.")
			return
		}

		liveBroadcast := &youtube.LiveBroadcast{
			Snippet: &youtube.LiveBroadcastSnippet{
				Title:              s.Title,
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

		if u.dryRun {
			logging.YLSLogger().Info("would have created LiveBroadcast resource, but is dry-run",
				zap.String("streamName", s.Name),
				zap.String("title", liveBroadcast.Snippet.Title),
				zap.String("description", liveBroadcast.Snippet.Description),
				zap.String("scheduledStart", liveBroadcast.Snippet.ScheduledStartTime),
				zap.String("privacyLevel", liveBroadcast.Status.PrivacyStatus),
			)
		} else {
			liveBroadcastCall := u.svc.LiveBroadcasts.Insert([]string{"snippet", "status", "content_details"}, liveBroadcast)
			broadcastResp, err := liveBroadcastCall.Do()
			if err != nil {
				logging.YLSLogger().Error("failed to create a live broadcast", zap.String("streamName", s.Name), zap.Error(err))
				return
			}

			liveStreamCall := u.svc.LiveStreams.List([]string{"cdn"}).Mine(true)
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
				zap.String("shareableLink", fmt.Sprintf("https://youtube.com/live/%s?feature=share", broadcastResp.Id)),
				zap.String("embedableLink", fmt.Sprintf("https://youtube.com/embed/%s", broadcastResp.Id)),
			)

			if s.Publisher != nil {
				p, err := s.Publisher.GetPublisher()
				if err != nil {
					logging.YLSLogger().Fatal("unable to publish using provided publisher config", zap.Error(err))
				}

				if err := p.Publish(broadcastResp, s); err != nil {
					logging.YLSLogger().Fatal("unable to publish Youtube Live Broadcast to publish target", zap.Error(err))
				}

				logging.YLSLogger().Info("published stream to publish target using configured publisher", zap.Any("publisherName", s.Publisher))
			} else {
				logging.YLSLogger().Warn("no publisher config specified for stream. skipping stream publish. don't worry, the Youtube livestream was still created",
					zap.Any("stream", s),
				)
			}
		}
	}
}

type StreamList struct {
	Items []Stream `yaml:"streams"`
}

type Stream struct {
	Name              string               `yaml:"name"`
	Title             string               `yaml:"title"`
	Description       string               `yaml:"description"`
	Schedule          string               `yaml:"schedule"`
	StartDelaySeconds uint16               `yaml:"delaySeconds"`
	PrivacyLevel      string               `yaml:"privacyLevel"`
	Publisher         *pub.PublisherConfig `yaml:"publisher,omitempty"`
}

func (s *Stream) String() string {
	return fmt.Sprintf(
		`{
			'name': %s,
			'title': %s,
			'description': %s,
			'schedule': %s,
			'startDelay': %q seconds,
			'privacyLevel': %s,
			'publisherName': %s,
		}`,
		s.Name,
		s.Title,
		s.Description,
		s.Schedule,
		s.StartDelaySeconds,
		s.PrivacyLevel,
		s.Publisher,
	)
}
