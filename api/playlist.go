package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zmb3/spotify/v2"
	"io/ioutil"
	"net/http"
)

type PlaylistItem struct {
	Uri        string `json:"uri"`
	Attributes struct {
		Timestamp        string `json:"timestamp"`
		FormatAttributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"formatAttributes"`
		ItemId string `json:"itemId"`
	} `json:"attributes"`
}

type PlaylistMetadata struct {
	Revision   string `json:"revision"`
	Length     int    `json:"length"`
	Attributes struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		Collaborative    bool   `json:"collaborative"`
		Format           string `json:"format"`
		FormatAttributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"formatAttributes"`
		PictureSize []struct {
			TargetName string `json:"targetName"`
			Url        string `json:"url"`
		} `json:"pictureSize"`
	} `json:"attributes"`
	Contents struct {
		Pos       int            `json:"pos"`
		Truncated bool           `json:"truncated"`
		Items     []PlaylistItem `json:"items"`
	} `json:"contents"`
	Timestamp             string `json:"timestamp"`
	OwnerUsername         string `json:"ownerUsername"`
	AbuseReportingEnabled bool   `json:"abuseReportingEnabled"`
	Capabilities          struct {
		CanView                    bool `json:"canView"`
		CanAdministratePermissions bool `json:"canAdministratePermissions"`
		CanEditMetadata            bool `json:"canEditMetadata"`
		CanEditItems               bool `json:"canEditItems"`
	} `json:"capabilities"`
}

func (c *Client) GetMetadata(ctx context.Context, playlistID spotify.ID) (*PlaylistMetadata, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"",
		fmt.Sprintf("https://spclient.wg.spotify.com/playlist/v2/playlist/%s?from=0&length=100", playlistID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))
	playlist := &PlaylistMetadata{}
	if err = json.Unmarshal(body, playlist); err != nil {
		return nil, err
	}

	return playlist, nil
}

func (c *Client) getSpotifiedPlaylist(ctx context.Context, search string) (*spotify.SimplePlaylist, error) {
	playlists, err := c.Client.Search(ctx, search, spotify.SearchTypePlaylist)
	if err != nil {
		return nil, err
	}

	for _, playlist := range playlists.Playlists.Playlists {
		if playlist.Owner.ID == "spotify" {
			return &playlist, nil
		}
	}

	return nil, nil
}

func (c *Client) GetDiscovery(ctx context.Context) (*spotify.SimplePlaylist, error) {
	return c.getSpotifiedPlaylist(ctx, "Discover Weekly")
}

func (c *Client) GetRelease(ctx context.Context) (*spotify.SimplePlaylist, error) {
	return c.getSpotifiedPlaylist(ctx, "Release Radar")
}

func isDislikedTrack(track *spotify.PlaylistTrack, metadata *PlaylistMetadata) bool {
	for _, item := range metadata.Contents.Items {
		if item.Uri != string(track.Track.URI) {
			continue
		}

		for _, attribute := range item.Attributes.FormatAttributes {
			if attribute.Key == "dislike-feedback-selected" && attribute.Value == "1" {
				return true
			}
		}
	}

	return false
}

func (c *Client) GetNotDislikedTracks(ctx context.Context, playlist *spotify.SimplePlaylist) ([]spotify.ID, []spotify.ID, error) {
	var result []spotify.ID
	var disliked []spotify.ID

	allTracks, err := c.Client.GetPlaylistTracks(ctx, playlist.ID, spotify.Limit(100))
	if err != nil {
		return result, disliked, err
	}

	metadata, err := c.GetMetadata(ctx, playlist.ID)
	if err != nil {
		return result, disliked, err
	}

	for _, track := range allTracks.Tracks {
		if isDislikedTrack(&track, metadata) {
			disliked = append(disliked, track.Track.ID)
		} else {
			result = append(result, track.Track.ID)
		}
	}

	return result, disliked, nil
}

func (c *Client) RemoveDuplicateTracks(ctx context.Context, playlistId spotify.ID, tracksToSave []spotify.ID) ([]spotify.ID, error) {
	offset := 0
	limit := 100
	var tracksInPlaylist []spotify.ID
	for {
		tracks, err := c.Client.GetPlaylistTracks(ctx, playlistId, spotify.Limit(limit), spotify.Offset(offset))

		if err != nil {
			return nil, err
		}

		for _, track := range tracks.Tracks {
			tracksInPlaylist = append(tracksInPlaylist, track.Track.ID)
		}

		offset += limit
		if tracks.Next == "" {
			break
		}
	}

	var uniqueTracks []spotify.ID
	for _, trackToSave := range tracksToSave {
		add := true
		for _, trackInPlaylist := range tracksInPlaylist {
			if trackToSave == trackInPlaylist {
				add = false
				break
			}
		}

		if add {
			uniqueTracks = append(uniqueTracks, trackToSave)
		}
	}

	return uniqueTracks, nil
}
