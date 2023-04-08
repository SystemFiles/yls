package pub

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

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
	TLS      bool   `yaml:"tls"`
	Username string `yaml:"username"`
	AppToken string `yaml:"appToken"`
	// Indexing preferences
	ExistingPageId int `yaml:"existingPageID,omitempty"`
	// Wordpress payload data
	Data WordpressData `yaml:"data"`
}

type WordpressMeta struct {
	Id             int    `yaml:"existingId,omitempty"`
	Type           string `yaml:"type,omitempty"`
	TitleOverride  string `yaml:"titleOverride,omitempty"`
	Slug           string `yaml:"slug,omitempty"`
	Password       string `yaml:"password,omitempty"`
	Status         string `yaml:"status,omitempty"`
	CommentStatus  string `yaml:"comment_status,omitempty"`
	Parent         int    `yaml:"parent,omitempty"`
	AuthorOverride int    `yaml:"author,omitempty"`
	// future fields (will not do anything right now)
	FeaturedImage int `yaml:"featured_image,omitempty"`
}

type WordpressData struct {
	Meta    WordpressMeta `yaml:"meta"`
	Content string        `yaml:"content"`
}

const (
	CONTENT_TYPE_BLOGPOST = "post"
	CONTENT_TYPE_PAGE     = "page"
)

var CONTENT_TYPES_ALLOWED = []string{CONTENT_TYPE_BLOGPOST, CONTENT_TYPE_PAGE}

func NewWordpressPublisher(cfg *WordpressConfig) (*Wordpress, error) {
	proto := "http"
	if cfg.TLS {
		proto = "https"
	}
	wpClient, err := cfg.getClient(
		fmt.Sprintf("%s://%s:%s/wp-json/wp/v2", proto, cfg.Host, cfg.Port),
		cfg.Username,
		cfg.AppToken,
	)
	if err != nil {
		return nil, err
	}

	return &Wordpress{
		client: wpClient,
		data:   &cfg.Data,
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

/*
WORDPRESS CLIENT OBJECT
*/
type Wordpress struct {
	data   *WordpressData
	client *wordpress.Client
}

func (w *Wordpress) templatePage(vars interface{}) (string, error) {
	var res bytes.Buffer
	tmpl, err := template.New("template").Funcs(sprig.FuncMap()).Parse(w.data.Content)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&res, vars)

	logging.YLSLogger().Debug("templated pagecontent for wordpress publisher",
		zap.String("content", res.String()),
	)
	return res.String(), err
}

func (w *Wordpress) Publish(broadcast *youtube.LiveBroadcast, publishVars interface{}) error {
	type Vars struct {
		Broadcast *youtube.LiveBroadcast
		ExtraVars interface{}
	}

	if !stringInSlice(defaultValue(w.data.Meta.Type, CONTENT_TYPE_PAGE, ""), CONTENT_TYPES_ALLOWED) {
		return fmt.Errorf("invalid value for Wordpress content type. must be one of [%s]", strings.Join(CONTENT_TYPES_ALLOWED, ", "))
	}

	pageContent, err := w.templatePage(&Vars{
		Broadcast: broadcast,
		ExtraVars: publishVars,
	})
	if err != nil {
		return err
	}

	post := &wordpress.Post{
		Password:      w.data.Meta.Password,
		Slug:          w.data.Meta.Slug,
		Status:        defaultValue(w.data.Meta.Status, wordpress.PostStatusPrivate, ""),
		Type:          defaultValue(w.data.Meta.Type, CONTENT_TYPE_PAGE, ""),
		Title:         wordpress.Title{Raw: broadcast.Snippet.Title},
		Content:       wordpress.Content{Raw: pageContent},
		Author:        w.data.Meta.AuthorOverride,
		CommentStatus: w.data.Meta.CommentStatus,
	}

	page := &wordpress.Page{
		Password:      w.data.Meta.Password,
		Slug:          w.data.Meta.Slug,
		Status:        defaultValue(w.data.Meta.Status, wordpress.PostStatusPrivate, ""),
		Type:          defaultValue(w.data.Meta.Type, CONTENT_TYPE_PAGE, ""),
		Title:         wordpress.Title{Raw: defaultValue(w.data.Meta.TitleOverride, broadcast.Snippet.Title, "")},
		Content:       wordpress.Content{Raw: pageContent},
		Author:        w.data.Meta.AuthorOverride,
		CommentStatus: w.data.Meta.CommentStatus,
	}

	// update the page using an existing ID reference
	if w.data.Meta.Id != 0 {
		logging.YLSLogger().Debug("updating existing page with new Stream",
			zap.String("content", pageContent),
			zap.Int("existingPageID", w.data.Meta.Id),
		)

		if w.data.Meta.Type == CONTENT_TYPE_BLOGPOST {
			_, _, _, err := w.client.Posts().Update(w.data.Meta.Id, post)
			return err
		}
		_, _, _, err := w.client.Pages().Update(w.data.Meta.Id, page)
		return err
	}

	logging.YLSLogger().Debug("creating new post for stream publish",
		zap.String("postTitle", broadcast.Snippet.Title),
		zap.String("content", pageContent),
	)
	if w.data.Meta.Type == CONTENT_TYPE_BLOGPOST {
		_, _, _, err := w.client.Posts().Create(post)
		return err
	}
	_, _, _, err = w.client.Pages().Create(page)
	return err
}
