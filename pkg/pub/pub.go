package pub

import (
	"google.golang.org/api/youtube/v3"
)

type PublishOptions struct {
	WPConfig string
	WPBaseURL string
	WPUsername string
	WPAppToken string
	WPPageTemplate string
	WPExistingPageId int
}

type Publisher interface {
	Publish(s *youtube.LiveBroadcast) error
	Configure(cmd *PublishOptions)
}
