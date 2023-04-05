package stream

import (
	"context"
	"errors"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/client"
	"sykesdev.ca/yls/pkg/logging"
	"sykesdev.ca/yls/pkg/pub"
)

type StreamSchedulerConfig struct {
	pubConfig *pub.PublisherConfig `yaml:"publisher,omitempty"`
	OAuthConfig  string               `yaml:"oauthConfig"`
	SecretsCache string               `yaml:"secretsCache"`
	streams       []*Stream `yaml:"stream"`
}

func (StreamSchedulerConfig) initYoutubeService(ctx context.Context, oauthConfig, secretsCache string) (*youtube.Service, error) {
	if oauthConfig == "" {
		return nil, errors.New("oauth configuration file is required. specify --oauth-config or use the environment variable YLS_OAUTH_CONFIG")
	}
	b, err := os.ReadFile(oauthConfig)
	if err != nil {
		logging.YLSLogger().Fatal("Unable to read oauth configuration from file", zap.Error(err))
	}

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

func (s *StreamSchedulerConfig) iterator() *StreamIterator {
	return &StreamIterator{
		streams: s.streams,
	}
}

type StreamScheduler struct {
	pub.Publisher
	Svc       *youtube.Service
	dryRun    bool
	Streams    *StreamIterator
}

func NewScheduler(ctx context.Context, cfg *StreamSchedulerConfig) (*StreamScheduler, error) {
	svc, err := cfg.initYoutubeService(context.Background(), cfg.OAuthConfig, cfg.SecretsCache)
	if err != nil {
		return nil, err
	}

	var p pub.Publisher
	if cfg.pubConfig != nil {
		p, err = cfg.pubConfig.GetPublisher()
		if err != nil {
			return nil, err
		}
	}

	if len(cfg.streams) == 0 {
		return nil, errors.New("must specify at least one stream")
	}

	return &StreamScheduler{
		Svc:       svc,
		Streams:    cfg.iterator(),
		dryRun:    false,
		Publisher: p,
	}, nil
}