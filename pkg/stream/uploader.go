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

func (u *StreamUploadClient) uploadThumbnail(s *youtube.ThumbnailsService, videoId, thumbnailPath string) (*youtube.ThumbnailSetResponse, error) {
	f, err := os.Open(thumbnailPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	logging.YLSLogger().Debug("thumbnail being uploaded", zap.String("path", thumbnailPath))
	ts := s.Set(videoId)

	resp, err := ts.Media(f).Do()
	if err != nil {
		return nil, err
	}

	logging.YLSLogger().Debug("thumbnail upload request completed!", zap.Int("status", resp.HTTPStatusCode))

	return resp, nil
}

func (u *StreamUploadClient) prepareThumbnails(s *Stream, b *youtube.LiveBroadcast) *youtube.ThumbnailDetails {
	var err error

	const T_SET_DEFAULT = "default"
	const T_SET_STANDARD = "standard"
	const T_SET_MEDIUM = "medium"
	const T_SET_HIGH = "high"
	const T_SET_MAXRES = "maxres"

	tSetResponses := make(map[string]*youtube.ThumbnailSetResponse, 5)
	videoId := b.Id
	thumbnailSvc := youtube.NewThumbnailsService(u.svc)

	// defaultUrl is a helper to create a consistent format for the returned ThumbnailDetails struct
	defaultUrl := func(r *youtube.ThumbnailSetResponse) string {
		if r != nil && len(r.Items) > 0 && r.Items[0] != nil {
			return r.Items[0].Default.Url
		}

		logging.YLSLogger().Debug("thumbnail set response was empty ... using default")
		return ""
	}

	// Default
	if s.Thumbnail.Default.Path != "" {
		tSetResponses[T_SET_DEFAULT], err = u.uploadThumbnail(thumbnailSvc, videoId, s.Thumbnail.Default.Path)
		if err != nil {
			logging.YLSLogger().Error("unable to upload thumbnail for live broadcast",
				zap.String("broadcastId", videoId),
				zap.String("thumbnail_type", T_SET_DEFAULT),
				zap.Error(err),
			)
		}
	}

	// Standard
	if s.Thumbnail.Standard.Path != "" {
		tSetResponses[T_SET_STANDARD], err = u.uploadThumbnail(thumbnailSvc, videoId, s.Thumbnail.Standard.Path)
		if err != nil {
			logging.YLSLogger().Error("unable to upload thumbnail for live broadcast",
				zap.String("broadcastId", videoId),
				zap.String("thumbnail_type", T_SET_STANDARD),
				zap.Error(err),
			)
		}
	}

	// Medium
	if s.Thumbnail.Medium.Path != "" {
		tSetResponses[T_SET_MEDIUM], err = u.uploadThumbnail(thumbnailSvc, videoId, s.Thumbnail.Medium.Path)
		if err != nil {
			logging.YLSLogger().Error("unable to upload thumbnail for live broadcast",
				zap.String("broadcastId", videoId),
				zap.String("thumbnail_type", T_SET_MEDIUM),
				zap.Error(err),
			)
		}
	}

	// High
	if s.Thumbnail.High.Path != "" {
		tSetResponses[T_SET_HIGH], err = u.uploadThumbnail(thumbnailSvc, videoId, s.Thumbnail.High.Path)
		if err != nil {
			logging.YLSLogger().Error("unable to upload thumbnail for live broadcast",
				zap.String("broadcastId", videoId),
				zap.String("thumbnail_type", T_SET_HIGH),
				zap.Error(err),
			)
		}
	}

	// Max Resolution
	if s.Thumbnail.Maxres.Path != "" {
		tSetResponses[T_SET_MAXRES], err = u.uploadThumbnail(thumbnailSvc, videoId, s.Thumbnail.Maxres.Path)
		if err != nil {
			logging.YLSLogger().Error("unable to upload thumbnail for live broadcast",
				zap.String("broadcastId", videoId),
				zap.String("thumbnail_type", T_SET_MAXRES),
				zap.Error(err),
			)
		}
	}

	return &youtube.ThumbnailDetails{
		Default: &youtube.Thumbnail{
			Width:  s.Thumbnail.Default.Width,
			Height: s.Thumbnail.Default.Height,
			Url:    defaultUrl(tSetResponses[T_SET_DEFAULT]),
		},
		High: &youtube.Thumbnail{
			Width:  s.Thumbnail.High.Width,
			Height: s.Thumbnail.High.Height,
			Url:    defaultUrl(tSetResponses[T_SET_HIGH]),
		},
		Maxres: &youtube.Thumbnail{
			Width:  s.Thumbnail.Maxres.Width,
			Height: s.Thumbnail.Maxres.Height,
			Url:    defaultUrl(tSetResponses[T_SET_MAXRES]),
		},
		Medium: &youtube.Thumbnail{
			Width:  s.Thumbnail.Medium.Width,
			Height: s.Thumbnail.Medium.Height,
			Url:    defaultUrl(tSetResponses[T_SET_MEDIUM]),
		},
		Standard: &youtube.Thumbnail{
			Width:  s.Thumbnail.Standard.Width,
			Height: s.Thumbnail.Standard.Height,
			Url:    defaultUrl(tSetResponses[T_SET_STANDARD]),
		},
	}

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
				PrivacyStatus:           s.Privacy.Level,
				SelfDeclaredMadeForKids: s.Privacy.SelfDeclaredMadeForKids,
			},
			ContentDetails: s.ContentDetails.Make(),
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

			// Upload and assign thumbnail to LiveBroadcast
			logging.YLSLogger().Info("assigning configured thumbnails to published LiveBroadcast")
			thumbnails := u.prepareThumbnails(s, broadcastResp)

			liveBroadcastUpdateCall := u.svc.LiveBroadcasts.Update([]string{"snippet"}, &youtube.LiveBroadcast{
				Id: broadcastResp.Id,
				Snippet: &youtube.LiveBroadcastSnippet{
					Title:              broadcastResp.Snippet.Title,
					Description:        broadcastResp.Snippet.Description,
					ScheduledStartTime: broadcastResp.Snippet.ScheduledStartTime,
					Thumbnails:         thumbnails,
				},
			})
			_, err = liveBroadcastUpdateCall.Do()
			if err != nil {
				logging.YLSLogger().Error("failed to update existing live broadcast with Thumbnail",
					zap.String("broadcastId", broadcastResp.Id),
					zap.String("streamName", s.Name),
					zap.Error(err),
				)
			}

			logging.YLSLogger().Info("uploaded and attached thumbnail to existing live broadcast successfully",
				zap.String("streamName", s.Name),
				zap.String("broadcastName", broadcastResp.Snippet.Title),
				zap.String("scheduledStart", broadcastResp.Snippet.ScheduledStartTime),
				zap.String("currentStatus", broadcastResp.Status.RecordingStatus),
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
