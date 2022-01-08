package main

import (
	"context"
	"github.com/hashicorp/go-multierror"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.uber.org/dig"
	"spotify/api"
	"spotify/config"
	"spotify/token"
)

func getConfigManager() *config.Manager {
	return config.NewManager("config.json")
}

func getConfig(manager *config.Manager) (*config.Config, error) {
	return manager.Get()
}

func getAuth(cfg *config.Config) *spotifyauth.Authenticator {
	return spotifyauth.New(
		spotifyauth.WithRedirectURL(cfg.RedirectUrl),
		spotifyauth.WithClientID(cfg.ClientId),
		spotifyauth.WithClientSecret(cfg.ClientSecret),
		spotifyauth.WithScopes(
			spotifyauth.ScopeImageUpload,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
			spotifyauth.ScopeUserFollowModify,
			spotifyauth.ScopeUserFollowRead,
			spotifyauth.ScopeUserLibraryModify,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserReadEmail,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserReadRecentlyPlayed,
			spotifyauth.ScopeUserTopRead,
			spotifyauth.ScopeStreaming,
		),
	)
}

func getTokenManager() *token.Manager {
	return token.NewManager("token.json")
}

func getApiClient(auth *spotifyauth.Authenticator, tokenManager *token.Manager) (*api.Client, error) {
	tok, err := tokenManager.GetToken()
	if err != nil {
		return nil, err
	}

	return api.New(auth, tok), nil
}

func getContainer() (*dig.Container, error) {
	c := dig.New()

	err := multierror.Append(
		nil,
		c.Provide(getConfigManager),
		c.Provide(getConfig),
		c.Provide(getApiClient),
		c.Provide(getAuth),
		c.Provide(getTokenManager),
		c.Provide(func() context.Context { return context.Background() }),
	)

	return c, err.ErrorOrNil()
}
