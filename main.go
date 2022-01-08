package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"spotify/api"
	"spotify/config"
)

func main() {
	err := run()

	if err != nil {
		panic(err)
	}
}

func run() error {
	if len(os.Args) == 1 {
		return fmt.Errorf("Give me a command!!!")
	}

	cmd := os.Args[1]

	c, err := getContainer()

	if err != nil {
		return err
	}

	if cmd == "init" {
		return c.Invoke(initConfigCommand)
	}

	switch cmd {
	case "discover":
		err = c.Invoke(saveDiscoverCommand)
		break
	case "release":
		err = c.Invoke(saveReleaseCommand)
		break
	case "auth":
		err = c.Invoke(authCommand)
		break
	default:
		return fmt.Errorf("There is no " + cmd + " command")
	}

	return err
}

func initConfigCommand(manager *config.Manager) error {
	link := "https://developer.spotify.com/dashboard/applications"
	fmt.Printf("Go to %s to create an app and retrieve credentials \n", link)

	return manager.Init()
}

func saveDiscoverCommand(ctx context.Context, cfg *config.Config, client *api.Client) error {
	playlist, err := client.GetDiscovery(ctx)
	if err != nil {
		return err
	}
	tracks, _, err := client.GetNotDislikedTracks(ctx, playlist)
	if err != nil {
		return err
	}

	err = saveToPlaylist(ctx, client, cfg.DiscoverTargetPlaylist, tracks)

	if err != nil {
		return err
	}

	tok, err := client.GetToken()
	if err != nil {
		return err
	}
	return getTokenManager().SaveToken(tok)
}

func saveReleaseCommand(ctx context.Context, cfg *config.Config, client *api.Client) error {
	playlist, err := client.GetRelease(ctx)
	if err != nil {
		return err
	}
	tracks, _, err := client.GetNotDislikedTracks(ctx, playlist)
	if err != nil {
		return err
	}

	err = saveToPlaylist(ctx, client, cfg.ReleaseTargetPlaylist, tracks)

	if err != nil {
		return err
	}
	tok, err := client.GetToken()
	if err != nil {
		return err
	}
	return getTokenManager().SaveToken(tok)
}

func saveToPlaylist(ctx context.Context, client *api.Client, playlistId spotify.ID, tracks []spotify.ID) error {
	uniqueTracks, err := client.RemoveDuplicateTracks(ctx, playlistId, tracks)
	if err != nil {
		return err
	}

	if len(uniqueTracks) == 0 {
		log.Println("No new tracks to add")
		return nil
	}

	_, err = client.AddTracksToPlaylist(ctx, playlistId, uniqueTracks...)
	if err != nil {
		return err
	}

	log.Printf("Added %d tracks\n", len(uniqueTracks))

	return nil
}

func authCommand(ctx context.Context, cfg *config.Config, auth *spotifyauth.Authenticator) error {
	mux := http.NewServeMux()
	ch := make(chan *oauth2.Token)
	state := uuid.NewString()

	// first start an HTTP server
	mux.HandleFunc("/callback", completeAuth(auth, state, ch))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	var srv *http.Server
	go func() {
		srv = &http.Server{Addr: cfg.ServerAddr, Handler: mux}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	tok := <-ch

	if err := getTokenManager().SaveToken(tok); err != nil {
		return err
	}

	return srv.Shutdown(ctx)
}

func completeAuth(auth *spotifyauth.Authenticator, state string, ch chan<- *oauth2.Token) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}

		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}

		w.Write([]byte("Auth complete"))

		ch <- tok
	}
}
