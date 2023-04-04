package pub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/client"
	"sykesdev.ca/yls/pkg/logging"
)

type YoutubeStreamConfigList struct {
	Streams []YoutubeStreamConfig `yaml:"streams"`
}

type YoutubeConfig struct {
	OAuthConfig  string               `yaml:"oauthConfig"`
	SecretsCache string               `yaml:"secretsCache"`
	Stream       *YoutubeStreamConfig `yaml:"stream"`
}

type YoutubeStreamConfig struct {
	Name              string `yaml:"name"`
	Title             string `yaml:"title"`
	Description       string `yaml:"description"`
	Schedule          string `yaml:"schedule"`
	StartDelaySeconds uint16 `yaml:"delaySeconds"`
	PrivacyLevel      string `yaml:"privacyLevel"`
}

func (YoutubeConfig) initYoutubeService(ctx context.Context, oauthConfig, secretsCache string) (*youtube.Service, error) {
	if oauthConfig == "" {
		return nil, errors.New("oauth configuration file is required. specify --oauth-config or use the environment variable YLS_OAUTH_CONFIG")
	}
	b, err := os.ReadFile(oauthConfig)
	if err != nil {
		logging.YLSLogger().Fatal("Unable to read oauth configuration from file", zap.Error(err))
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		logging.YLSLogger().Fatal("Unable to parse client secret file to config", zap.Error(err))
	}

	client := client.Get(ctx, secretsCache, config)
	svc, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func NewYoutubePublisher(cfg *YoutubeConfig) (*Youtube, error) {
	svc, err := cfg.initYoutubeService(context.Background(), cfg.OAuthConfig, cfg.SecretsCache)
	if err != nil {
		return nil, err
	}

	return &Youtube{
		svc:       svc,
		stream:    cfg.Stream,
		publisher: nil,
		dryRun:    false,
	}, nil
}

type Youtube struct {
	svc       *youtube.Service
	stream    *YoutubeStreamConfig
	publisher Publisher
	dryRun    bool
}

func (y *Youtube) DryRun() *Youtube {
	y.dryRun = true
	return y
}

func (y *Youtube) WithPublisher(pub Publisher) *Youtube {
	y.publisher = pub
	return y
}

func (y *Youtube) Publish(s *YoutubeStreamConfig, b *youtube.LiveBroadcast) error {
	if y.dryRun {
		logging.YLSLogger().Info("would have created LiveBroadcast resource, but is dry-run",
			zap.String("streamName", s.Name),
			zap.String("title", b.Snippet.Title),
			zap.String("description", b.Snippet.Description),
			zap.String("scheduledStart", b.Snippet.ScheduledStartTime),
			zap.String("privacyLevel", b.Status.PrivacyStatus),
		)
	} else {
		liveBroadcastCall := y.svc.LiveBroadcasts.Insert([]string{"snippet", "status", "content_details"}, b)
		broadcastResp, err := liveBroadcastCall.Do()
		if err != nil {
			return err
		}

		liveStreamCall := y.svc.LiveStreams.List([]string{"cdn"}).Mine(true)
		existingLiveStreams, err := liveStreamCall.Do()
		if err != nil {
			return err
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

		if y.publisher != nil {
			if err := y.publisher.Publish(s, broadcastResp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (y *Youtube) Go() {
	liveBroadcast := &youtube.LiveBroadcast{
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              y.stream.Title,
			Description:        y.stream.Description,
			ScheduledStartTime: time.Now().Local().Add(time.Duration(y.stream.StartDelaySeconds) * time.Second).Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus:           y.stream.PrivacyLevel,
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

	if err := y.Publish(y.stream, liveBroadcast); err != nil {
		logging.YLSLogger().Error("failed to publish livestream to Youtube using Youtube Publisher", zap.Error(err))
	}
}

// type StreamList struct {
// 	Items []Stream `yaml:"streams"`
// }

// type Stream struct {
// 	Name              string `yaml:"name"`
// 	Title             string `yaml:"title"`
// 	Description       string `yaml:"description"`
// 	Schedule          string `yaml:"schedule"`
// 	StartDelaySeconds uint16 `yaml:"delaySeconds"`
// 	PrivacyLevel      string `yaml:"privacyLevel"`

// 	dryRun    bool             `yaml:"-"`
// 	service   *youtube.Service `yaml:"-"`
// 	publisher Publisher    `yaml:"-"`
// }

// func (s *Stream) WithService(y *youtube.Service) *Stream {
// 	s.service = y
// 	return s
// }

// func (s *Stream) DryRun(val bool) *Stream {
// 	s.dryRun = val
// 	return s
// }

// func (s *Stream) WithPublisher(pub Publisher) *Stream {
// 	s.publisher = pub
// 	return s
// }

// func (s *Stream) Go() {
// 	if s.service == nil {
// 		logging.YLSLogger().Error("unable to create Live Broadcast resource. no service was available. make sure to chain `WithService(x)` before running `Go()`")
// 	}

// 	liveBroadcast := &youtube.LiveBroadcast{
// 		Snippet: &youtube.LiveBroadcastSnippet{
// 			Title:              s.Title,
// 			Description:        s.Description,
// 			ScheduledStartTime: time.Now().Local().Add(time.Duration(s.StartDelaySeconds) * time.Second).Format(time.RFC3339),
// 		},
// 		Status: &youtube.LiveBroadcastStatus{
// 			PrivacyStatus:           s.PrivacyLevel,
// 			MadeForKids:             false,
// 			SelfDeclaredMadeForKids: false,
// 		},
// 		ContentDetails: &youtube.LiveBroadcastContentDetails{
// 			EnableClosedCaptions: false,
// 			EnableAutoStop:       true,
// 			EnableDvr:            true,
// 			ClosedCaptionsType:   "closedCaptionsDisabled",
// 		},
// 	}

// 	if s.dryRun {
// 		logging.YLSLogger().Info("would have created LiveBroadcast resource, but is dry-run",
// 			zap.String("streamName", s.Name),
// 			zap.String("title", liveBroadcast.Snippet.Title),
// 			zap.String("description", liveBroadcast.Snippet.Description),
// 			zap.String("scheduledStart", liveBroadcast.Snippet.ScheduledStartTime),
// 			zap.String("privacyLevel", liveBroadcast.Status.PrivacyStatus),
// 		)
// 	} else {
// 		liveBroadcastCall := s.service.LiveBroadcasts.Insert([]string{"snippet", "status", "content_details"}, liveBroadcast)
// 		broadcastResp, err := liveBroadcastCall.Do()
// 		if err != nil {
// 			logging.YLSLogger().Error("failed to create a live broadcast", zap.String("streamName", s.Name), zap.Error(err))
// 			return
// 		}

// 		liveStreamCall := s.service.LiveStreams.List([]string{"cdn"}).Mine(true)
// 		existingLiveStreams, err := liveStreamCall.Do()
// 		if err != nil {
// 			logging.YLSLogger().Error("failed to get owned streams",
// 				zap.String("streamName", s.Name),
// 				zap.Error(err),
// 			)
// 		}

// 		streamKeys := []string{}
// 		for _, ls := range existingLiveStreams.Items {
// 			streamKeys = append(streamKeys, ls.Cdn.IngestionInfo.StreamName)
// 		}

// 		// Unfortunately it seems that the API cannot determine which Stream key will be used until a LS and Broadcast are bound.
// 		// Manually creating both can work, but will require a lot of extra detail in the configuration object which isn't ideal for users
// 		// For now just keep in mind: https://stackoverflow.com/questions/33798901/youtube-api-v3-get-live-now-rtmp-and-streamkey
// 		logging.YLSLogger().Info("created live scheduled broadcast",
// 			zap.String("streamName", s.Name),
// 			zap.String("broadcastName", broadcastResp.Snippet.Title),
// 			zap.String("scheduledStart", broadcastResp.Snippet.ScheduledStartTime),
// 			zap.String("currentStatus", broadcastResp.Status.RecordingStatus),
// 			zap.Strings("validStreamKeys", streamKeys),
// 			zap.String("shareableLink", fmt.Sprintf("https://youtube.com/live/%s?feature=share", broadcastResp.Id)),
// 			zap.String("embedableLink", fmt.Sprintf("https://youtube.com/embed/%s", broadcastResp.Id)),
// 		)

// 		if s.publisher != nil {
// 			if err := s.publisher.Publish(s, broadcastResp); err != nil {
// 				logging.YLSLogger().Fatal("unable to publish Youtube Live Broadcast to publish target", zap.Error(err))
// 			}
// 		}
// 	}
// }

// func (s *Stream) String() string {
// 	return s.Name
// }
