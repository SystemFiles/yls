package pub

import (
	"fmt"

	"google.golang.org/api/youtube/v3"
)

const PUBLISHER_WORDPRESS string = "wordpress"

type Publisher interface {
	Publish(broadcast *youtube.LiveBroadcast, publishVars interface{}) error
}

type PublisherConfig struct {
	Wordpress *WordpressConfig `yaml:"wordpress"`
}

func (p *PublisherConfig) GetPublisher() (Publisher, error) {
	if p.Wordpress != nil {
		return NewWordpressPublisher(p.Wordpress)
	}

	return nil, fmt.Errorf("unknown publisher")
}

func (p *PublisherConfig) String() string {
	if p.Wordpress != nil {
		return "wordpress"
	}

	return "unknown"
}
