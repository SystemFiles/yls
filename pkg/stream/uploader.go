package stream

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

type StreamUploaderConfig struct {
	Context     context.Context
	OauthConfig string
	Cache       string
	Scopes      []string
	DryRunMode  bool
}

type StreamUploadClient struct {
	svc    *youtube.Service
	dryRun bool
}

func New(cfg *StreamUploaderConfig) (*StreamUploadClient, error) {
	if cfg.OauthConfig == "" {
		return nil, errors.New("oauth configuration file is required. specify --oauth-config")
	}
	c, err := os.ReadFile(cfg.OauthConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to read the provided Oauth2 configuration file. %e", err)
	}

	config, err := google.ConfigFromJSON(c, cfg.Scopes...)
	if err != nil {
		return nil, err
	}

	client := client.Get(cfg.Context, cfg.Cache, config)
	svc, err := youtube.NewService(cfg.Context, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &StreamUploadClient{
		svc:    svc,
		dryRun: cfg.DryRunMode,
	}, nil
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
					zap.Any("stream", s.Name),
				)
			}
		}
	}
}
