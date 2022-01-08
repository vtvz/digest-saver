package api

import (
	"context"
	"fmt"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"net/http"
)

type Client struct {
	*spotify.Client
	Http *http.Client
}

func New(auth *spotifyauth.Authenticator, token *oauth2.Token) *Client {
	h := auth.Client(context.Background(), token)

	s := spotify.New(h)

	return &Client{s, h}
}

func (c *Client) GetToken() (*oauth2.Token, error) {
	oauthTransport, ok := c.Http.Transport.(*oauth2.Transport)
	if !ok {
		return nil, fmt.Errorf("cannot retrieve token")
	}

	return oauthTransport.Source.Token()
}
