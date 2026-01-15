package gcal

import "golang.org/x/oauth2"

type TokenStore interface {
	Load(userID string) (*oauth2.Token, error)
	Save(userID string, token *oauth2.Token) error
}
