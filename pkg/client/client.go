package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"sykesdev.ca/yls/pkg/logging"
)

var YLSLogger = logging.YLSLogger

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
		logging.YLSLogger().Warn("unable to cache oauth token", zap.Error(err))
	}
	defer f.Close()

	json.NewEncoder(f).Encode(token)
}
