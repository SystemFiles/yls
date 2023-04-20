package stream

import (
	"fmt"

	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/pub"
)

type StreamList struct {
	Items []Stream `yaml:"streams"`
}

type Stream struct {
	Name              string                        `yaml:"name"`
	Title             string                        `yaml:"title"`
	Thumbnail         *StreamThumbnailDetailsConfig `yaml:"thumbnails,omitempty"`
	Description       string                        `yaml:"description"`
	Schedule          string                        `yaml:"schedule"`
	StartDelaySeconds uint16                        `yaml:"delaySeconds"`
	Privacy           *StreamPrivacy                `yaml:"privacy,omitempty"`
	ContentDetails    StreamContentDetailsConfig    `yaml:"contentDetails,omitempty"`
	Publisher         *pub.PublisherConfig          `yaml:"publisher,omitempty"`
}

type StreamPrivacy struct {
	Level                   string `yaml:"level,omitempty"`
	MadeForKids             bool   `yaml:"madeForKids,omitempty"`
	SelfDeclaredMadeForKids bool   `yaml:"selfDeclaredMadeForKids,omitempty"`
}

type StreamContentDetailsConfig struct {
	ClosedCaptionsType      string `yaml:"closedCaptionsType,omitempty"`
	EnableAutoStart         bool   `yaml:"enableAutoStart,omitempty"`
	EnableAutoStop          bool   `yaml:"enableAutoStop,omitempty"`
	EnableClosedCaptions    bool   `yaml:"enableClosedCaptions,omitempty"`
	EnableContentEncryption bool   `yaml:"enableContentEncryption,omitempty"`
	EnableDvr               bool   `yaml:"enableDvr,omitempty"`
	EnableEmbed             bool   `yaml:"enableEmbed,omitempty"`
	EnableLowLatency        bool   `yaml:"enableLowLatency,omitempty"`
	LatencyPreference       string `yaml:"latencyPreference,omitempty"`
	Mesh                    string `yaml:"mesh,omitempty"`
	Projection              string `yaml:"projection,omitempty"`
	RecordFromStart         bool   `yaml:"recordFromStart,omitempty"`
	StartWithSlate          bool   `yaml:"startWithSlate,omitempty"`
	StereoLayout            string `yaml:"stereoLayout,omitempty"`
}

func (cd *StreamContentDetailsConfig) Make() *youtube.LiveBroadcastContentDetails {
	return &youtube.LiveBroadcastContentDetails{
		ClosedCaptionsType:      cd.ClosedCaptionsType,
		EnableAutoStart:         cd.EnableAutoStart,
		EnableAutoStop:          cd.EnableAutoStop,
		EnableClosedCaptions:    cd.EnableClosedCaptions,
		EnableContentEncryption: cd.EnableContentEncryption,
		EnableDvr:               cd.EnableDvr,
		EnableEmbed:             cd.EnableEmbed,
		EnableLowLatency:        cd.EnableLowLatency,
		LatencyPreference:       cd.LatencyPreference,
		Mesh:                    cd.Mesh,
		Projection:              cd.Projection,
		RecordFromStart:         cd.RecordFromStart,
		StartWithSlate:          cd.StartWithSlate,
		StereoLayout:            cd.StereoLayout,
	}
}

type StreamThumbnailDetailsConfig struct {
	Default  *StreamThumbnailConfig `yaml:"default,omitempty"`
	High     *StreamThumbnailConfig `yaml:"high,omitempty"`
	Maxres   *StreamThumbnailConfig `yaml:"maxres,omitempty"`
	Medium   *StreamThumbnailConfig `yaml:"medium,omitempty"`
	Standard *StreamThumbnailConfig `yaml:"standard,omitempty"`
}

func (td *StreamThumbnailDetailsConfig) Make() *youtube.ThumbnailDetails {
	return &youtube.ThumbnailDetails{
		Default:  td.Default.Make(),
		High:     td.High.Make(),
		Maxres:   td.Maxres.Make(),
		Medium:   td.Medium.Make(),
		Standard: td.Standard.Make(),
	}
}

type StreamThumbnailConfig struct {
	Width  int64  `yaml:"width,omitempty"`
	Height int64  `yaml:"height,omitempty"`
	Url    string `yaml:"url"`
}

func (t *StreamThumbnailConfig) Make() *youtube.Thumbnail {
	return &youtube.Thumbnail{
		Width:  t.Width,
		Height: t.Height,
		Url:    t.Url,
	}
}

type StreamThumbnail struct {
	youtube.ThumbnailDetails
}

func (s *Stream) String() string {
	return fmt.Sprintf("{'name': %s, 'title': %s, 'description': %s, 'schedule': %s, 'startDelay': %q seconds, 'privacyLevel': %s, 'publisherName': %s}",
		s.Name,
		s.Title,
		s.Description,
		s.Schedule,
		s.StartDelaySeconds,
		s.Privacy.Level,
		s.Publisher,
	)
}
