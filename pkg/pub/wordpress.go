package pub

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"regexp"

	"github.com/sogko/go-wordpress"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/yaml.v3"

	"sykesdev.ca/yls/pkg/logging"
)

type PublishVars struct {
	StreamURLEmbed string
	StreamURLShare string
}

type WordPressPublisher struct {
	// BaseURL is the Base URL for the wordpress v2 API
	BaseURL string `yaml:"apiBaseURL"`
	// Username is the Username for the user/service-account from which requests can be issued
	Username string `yaml:"username"`
	// AppToken is a generated revokable password used to authenticate the user provided by `username`
	// Note: the AppToken IS NOT the users password and is actually more similar to an API Key
	AppToken string `yaml:"appToken"`
	// ExistingPageId is an optional field where an optional wordpress Page ID can be specified for update
	// if existingpageId is not specified, a new page will be created for each publish operation
	ExistingPageId int `yaml:"existingPageID,omitempty"`
	// PageTemplate is a string that contains a gotemplate-compatible HTML page template that can be used to construct
	// a resulting page for a wordpress publish
	PageTemplate string `yaml:"pageTemplate"`
	// client represents a Go-Wordpress client that wraps the wordpress API into a more usable form
	// Note: cannot be marshalled/unmarshalled
	client *wordpress.Client `yaml:"-"`
}

func NewWordpressPublisher() *WordPressPublisher {
	return &WordPressPublisher{}
}

func configFromFile(p string) (*WordPressPublisher, error) {
	if p == "" {
		return nil, errors.New("must specify a valid filepath for wordpress publisher")
	}

	b, err := os.ReadFile(path.Clean(p))
	if err != nil {
		return nil, err
	}

	var wp WordPressPublisher
	err = yaml.Unmarshal(b, &wp)
	if err != nil {
		return nil, err
	}

	logging.YLSLogger().Debug("created Wordpress Publisher from config file",
		zap.Any("wordpressPublisher", wp),
	)
	return &wp, nil
}

func (w *WordPressPublisher) getClient() error {
	logging.YLSLogger().Debug("facts", zap.Any("pub", w))
	w.client = wordpress.NewClient(&wordpress.Options{
		BaseAPIURL: w.BaseURL,
		Username:   w.Username,
		Password:   w.AppToken,
	})

	_, _, _, err := w.client.Users().Me(nil)
	if err != nil {
		return fmt.Errorf("failed to instantiate a wordpress client. %e", err)
	}

	logging.YLSLogger().Debug("created wordpress client for publisher",
		zap.Any("client", w),
	)
	return nil
}

func (w *WordPressPublisher) templatePage(vars *PublishVars) (string, error) {
	var res bytes.Buffer
	tmpl, err := template.New("p").Parse(w.PageTemplate)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&res, vars)

	logging.YLSLogger().Debug("templated pagecontent for wordpress publisher",
		zap.String("content", res.String()),
	)
	return res.String(), err
}

func (w *WordPressPublisher) Publish(b *youtube.LiveBroadcast) error {
	err := w.getClient()
	if err != nil {
		return err
	}
	c := w.client

	pVars := &PublishVars{
		StreamURLEmbed: fmt.Sprintf("https://youtube.com/embed/%s?autoplay=0&livemonitor=1", b.Id),
		StreamURLShare: fmt.Sprintf("https://youtube.com/live/%s?feature=share", b.Id),
	}

	logging.YLSLogger().Debug("creating page content using publish vars",
		zap.Any("publishVars", pVars),
	)

	pageContent, err := w.templatePage(pVars)
	if err != nil {
		return err
	}

	if w.ExistingPageId != 0 {
		logging.YLSLogger().Debug("updating existing page with new Stream",
			zap.String("content", pageContent),
			zap.Int("existingPageID", w.ExistingPageId),
		)
		_, _, _, err := c.Pages().Update(w.ExistingPageId, &wordpress.Page{
			Content: wordpress.Content{
				Raw: pageContent,
			},
		})
		return err
	}

	logging.YLSLogger().Debug("creating new page for stream publish",
		zap.String("pageTitle", b.Snippet.Title),
		zap.String("content", pageContent),
	)
	_, _, _, err = c.Pages().Create(&wordpress.Page{
		Title: wordpress.Title{
			Raw: b.Snippet.Title,
		},
		Content: wordpress.Content{
			Raw: pageContent,
		},
	})
	return err
}

func (w *WordPressPublisher) Configure(cmd *PublishOptions) {
	p := &WordPressPublisher{}
	if cmd.WPConfig != "" {
		fp, err := configFromFile(cmd.WPConfig)
		if err != nil {
			logging.YLSLogger().Fatal("unable to construct publisher from config", zap.Error(err))
		}
		p = fp
	}

	if cmd.WPBaseURL != "" {
		p.BaseURL = cmd.WPBaseURL
	}

	if cmd.WPUsername != "" {
		p.Username = cmd.WPUsername
	}

	if cmd.WPAppToken != "" {
		p.AppToken = cmd.WPAppToken
	}

	if cmd.WPPageTemplate != "" {
		p.PageTemplate = cmd.WPPageTemplate
	}

	if cmd.WPExistingPageId != 0 {
		p.ExistingPageId = cmd.WPExistingPageId
	}

	if match, err := regexp.MatchString(`https?:\/\/.+`, p.BaseURL); err != nil || !match {
		logging.YLSLogger().Fatal("unable to construct publisher without a valid Base URL",
			zap.String("BaseURL", p.BaseURL),
		)
	}

	if p.Username == "" {
		logging.YLSLogger().Fatal("unable to construct publisher without a username")
	}

	if p.AppToken == "" {
		logging.YLSLogger().Fatal("unable to construct publisher without a password or app token")
	}

	if p.PageTemplate == "" {
		logging.YLSLogger().Fatal("unable to construct publisher without a page template")
	}

	logging.YLSLogger().Debug("FACTS", zap.Any("publisher", p))

	*w = WordPressPublisher{
		BaseURL:        p.BaseURL,
		Username:       p.Username,
		AppToken:       p.AppToken,
		ExistingPageId: p.ExistingPageId,
		PageTemplate:   p.PageTemplate,
		client:         &wordpress.Client{},
	}

	logging.YLSLogger().Debug("FACTS", zap.Any("publisher", p), zap.Any("ref", w))
}
