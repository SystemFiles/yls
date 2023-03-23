package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/youtube/v3"
	"sykesdev.ca/yls/pkg/logging"
)

var YLSLogger = logging.YLSLogger

type HeadlessCredentials struct {
	ClientSecretID    string `json:""`
	ClientSecret      string `json:"private_key"`
	ClientEmail       string `json:"client_email"`
	ClientID          string `json:"client_id"`
	Type              string `json:"type"`
	ProjectID         string `json:"project_id"`
	ClientX509CertURL string `json:"client_x509_cert_url"`
	TokenURI          string `json:"token_uri"`
}

func Get(ctx context.Context, secretsCache string, config *oauth2.Config) *http.Client {
	tok, err := tokenFromFile(secretsCache)
	if err != nil {
		YLSLogger().Warn("unable to get token from secrets file", zap.Error(err))
		YLSLogger().Warn("trying to get token using device-code flow instead ...")
		tok = getTokenFromWeb(ctx, config)
		saveToken(secretsCache, tok)
	}

	YLSLogger().Debug("successfully obtained access and refresh tokens for oauth client", zap.String("cacheLocation", secretsCache))
	return config.Client(ctx, tok)
}

func GetOauth2TokenSource(ctx context.Context, subject string, credentialsJSON []byte) (oauth2.TokenSource, error) {
	var ts oauth2.TokenSource
	gts, err := idtoken.NewTokenSource(ctx, "https://oauth2.googleapis.com/token",
		idtoken.WithCredentialsJSON(credentialsJSON),
		idtoken.WithCustomClaims(map[string]interface{}{"sub": "bensykes12@gmail.com"}),
	)
	// gts, err := google.JWTConfigFromJSON(credentialsJSON, youtube.YoutubeScope)
	if err != nil {
		return nil, err
	}
	ts = oauth2.ReuseTokenSource(nil, gts)

	tok, e := ts.Token()
	logging.YLSLogger().Info("TOKEN SORUUCAUIWHDIUHA", zap.Any("token", tok), zap.Error(e))

	return ts, nil
}

func GetIDTokenTokenSource(ctx context.Context, credentialsJSON []byte, subject string) (oauth2.TokenSource, error) {
	var ts oauth2.TokenSource
	gts, err := google.JWTConfigFromJSON(credentialsJSON, subject)
	if err != nil {
		return nil, err
	}
	ts = oauth2.ReuseTokenSource(nil, &idTokenSource{TokenSource: gts.TokenSource(ctx)})

	tok, e := ts.Token()
	logging.YLSLogger().Info("TOKEN SORUUCAUIWHDIUHA", zap.Any("token", tok), zap.Error(e))

	return ts, nil
}

// idTokenSource is an oauth2.TokenSource that wraps another
// It takes the id_token from TokenSource and passes that on as a bearer token
type idTokenSource struct {
	TokenSource oauth2.TokenSource
}

func (s *idTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("token did not contain an id_token")
	}

	return &oauth2.Token{
		AccessToken: idToken,
		TokenType:   "Bearer",
		Expiry:      token.Expiry,
	}, nil
}

func GetHeadlessTokenSource(ctx context.Context, subject string, credentialsJSON []byte) oauth2.TokenSource {
	var creds HeadlessCredentials
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		logging.YLSLogger().Fatal("unable to parse credentials for headless OAuth2.0 authentication from file contents", zap.Error(err))
	}

	jwtConf := jwt.Config{
		Email:         creds.ClientEmail,
		PrivateKey:    []byte(creds.ClientSecret),
		PrivateKeyID:  creds.ClientSecretID,
		Subject:       "sub: subject",
		Scopes:        []string{youtube.YoutubeScope},
		TokenURL:      creds.TokenURI,
		Expires:       0,
		Audience:      "",
		PrivateClaims: map[string]interface{}{},
		UseIDToken:    false,
	}

	token, e := jwtConf.TokenSource(ctx).Token()
	logging.YLSLogger().Info("token source token", zap.Any("token", token), zap.Error(e))

	return jwtConf.TokenSource(ctx)
}

func GetHeadless(ctx context.Context, credentialsJSON []byte) *http.Client {
	var creds HeadlessCredentials
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		logging.YLSLogger().Fatal("unable to parse credentials for headless OAuth2.0 authentication from file contents", zap.Error(err))
	}

	config := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeScope},
	}

	var ts *oauth2.Token
	tokenSource := config.TokenSource(ctx, ts)

	return oauth2.NewClient(ctx, tokenSource)
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to: \n%s\n", authURL)
	fmt.Println("Enter Authorization Code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		logging.YLSLogger().Fatal("unable to read authorization code from stdin", zap.Error(err))
	}

	tok, err := config.Exchange(ctx, code)
	if err != nil {
		logging.YLSLogger().Fatal("unable to retrieve token from web", zap.Error(err))
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	logging.YLSLogger().Debug("saving credentials to secrets cache", zap.String("cache_location", file))
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logging.YLSLogger().Fatal("unable to cache oauth token", zap.Error(err))
	}
	defer f.Close()

	json.NewEncoder(f).Encode(token)
}
