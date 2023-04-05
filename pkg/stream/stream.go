package stream

import (
	"fmt"

	"sykesdev.ca/yls/pkg/pub"
)

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
