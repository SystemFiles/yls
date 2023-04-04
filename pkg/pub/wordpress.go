package pub

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
	"github.com/sogko/go-wordpress"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/logging"
)

type WordpressConfig struct {
	// Connection
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	AppToken string `yaml:"appToken"`
	// Indexing preferences
	ExistingPageId int    `yaml:"existingPageID,omitempty"`
	Content        string `yaml:"content"`
}

func NewWordpressPublisher(cfg *WordpressConfig) (*Wordpress, error) {
	wpClient, err := cfg.getClient(
		fmt.Sprintf("https://%s:%s/wp-json/wp/v2", cfg.Host, cfg.Port),
		cfg.Username,
		cfg.AppToken,
	)
	if err != nil {
		return nil, err
	}

	return &Wordpress{
		pageID:  cfg.ExistingPageId,
		content: cfg.Content,
		client:  wpClient,
	}, nil
}

func (WordpressConfig) getClient(baseUrl, username, appToken string) (*wordpress.Client, error) {
	client := wordpress.NewClient(&wordpress.Options{
		BaseAPIURL: baseUrl,
		Username:   username,
		Password:   appToken,
	})

	_, _, _, err := client.Users().Me(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a wordpress client. %e", err)
	}

	logging.YLSLogger().Debug("created wordpress client for publisher",
		zap.Any("client", client),
	)
	return client, nil
}

type Wordpress struct {
	pageID  int
	content string
	client  *wordpress.Client
}

func (w *Wordpress) templatePage(vars interface{}) (string, error) {
	var res bytes.Buffer
	tmpl, err := template.New("template").Funcs(sprig.FuncMap()).Parse(w.content)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&res, vars)

	logging.YLSLogger().Debug("templated pagecontent for wordpress publisher",
		zap.String("content", res.String()),
	)
	return res.String(), err
}

func (w *Wordpress) Publish(stream *YoutubeStreamConfig, broadcast *youtube.LiveBroadcast) error {
	type vars struct {
		*YoutubeStreamConfig
		*youtube.LiveBroadcast
	}

	combVars := &vars{
		YoutubeStreamConfig: stream,
		LiveBroadcast:       broadcast,
	}

	pageContent, err := w.templatePage(combVars)
	if err != nil {
		return err
	}

	logging.YLSLogger().Debug("creating new page for stream publish",
		zap.String("pageTitle", broadcast.Snippet.Title),
		zap.String("content", pageContent),
	)
	_, _, _, err = w.client.Pages().Create(&wordpress.Page{
		Title: wordpress.Title{
			Raw: broadcast.Snippet.Title,
		},
		Content: wordpress.Content{
			Raw: pageContent,
		},
	})
	return err
}
