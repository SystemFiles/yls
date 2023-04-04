package export

import "sykesdev.ca/yls/pkg/pub"

type Registry interface {
	SendStream(string, pub.YoutubeStreamConfig)
	Register(string, pub.Publisher)
	Close()
}
