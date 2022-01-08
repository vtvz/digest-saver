package config

import "github.com/zmb3/spotify/v2"

type Config struct {
	ServerAddr             string     `json:"server_addr"`
	ClientId               string     `json:"client_id"`
	ClientSecret           string     `json:"client_secret"`
	RedirectUrl            string     `json:"redirect_url"`
	ReleaseTargetPlaylist  spotify.ID `json:"release_target_playlist"`
	DiscoverTargetPlaylist spotify.ID `json:"discover_target_playlist"`
}
