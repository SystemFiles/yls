package pub

import (
	"fmt"

	"google.golang.org/api/youtube/v3"
)

const PUBLISHER_WORDPRESS string = "wordpress"

type Publisher interface {
	Publish(*YoutubeStreamConfig, *youtube.LiveBroadcast) error
}

type PublisherConfig struct {
	Name      string           `yaml:"name"`
	Wordpress *WordpressConfig `yaml:"wordpress"`
}

func (p *PublisherConfig) GetPublisher() (Publisher, error) {
	if p.Wordpress != nil {
		return NewWordpressPublisher(p.Wordpress)
	}

	return nil, fmt.Errorf("unknown publisher")
}
