package gcal

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

type LoginConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	State        string
}

func LoginURL(cfg LoginConfig) string {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}
	return oauthCfg.AuthCodeURL(
		cfg.State,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

type ExchangeConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Code         string
}

func Exchange(ctx context.Context, cfg ExchangeConfig) (*oauth2.Token, error) {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}
	return oauthCfg.Exchange(ctx, cfg.Code)
}
