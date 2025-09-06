package gcal

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	UserID       string
	TokenStore   TokenStore
}

type Client struct {
	svc *calendar.Service
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	token, err := cfg.TokenStore.Load(cfg.UserID)
	if err != nil {
		return nil, ErrNotAuthorized
	}

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	ts := oauthCfg.TokenSource(ctx, token)
	httpClient := oauth2.NewClient(ctx, ts)

	svc, err := calendar.NewService(
		ctx,
		option.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	return &Client{svc: svc}, nil
}
