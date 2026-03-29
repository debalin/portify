package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/youtube"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	yt "google.golang.org/api/youtube/v3"
)

const redirectURI = "http://127.0.0.1:8080/callback"

var (
	ch          = make(chan *oauth2.Token)
	state       = "some-random-state-string"
	oauthConfig *oauth2.Config
)

func main() {
	// Load the .env file if it exists
	if err := godotenv.Load("../../.env"); err != nil {
		// Also try looking in current directory depending on where go run is executed from
		_ = godotenv.Load(".env")
	}

	clientID := os.Getenv("YOUTUBE_ID")
	clientSecret := os.Getenv("YOUTUBE_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("You must set YOUTUBE_ID and YOUTUBE_SECRET environment variables. Get them from Google Cloud Console.")
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURI,
		Scopes:       []string{yt.YoutubeScope},
	}

	// Start local server to receive callback
	http.HandleFunc("/callback", completeAuth)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println("Please log in to Google to grant YouTube access:")
	fmt.Println(url)
	openBrowser(url)

	fmt.Println("Waiting for authorization...")
	token := <-ch
	fmt.Println("Authorization successful! Token received.")

	// Create a dummy playlist using our CanonicalModel
	ctx := context.Background()
	dummyPlaylist := &converterv1.CanonicalPlaylist{
		Name:        "Test Playlist from Converter",
		Description: "This playlist was generated automatically to test the YouTube Adapter.",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Lofi Girl", Artist: "Lofi Girl"}, // Will search YouTube for this query
			{Title: "Chillhop", Artist: "Chillhop Music"},
		},
	}

	adapter := youtube.NewAdapter()
	
	fmt.Printf("\nCreating playlist \"%s\" on YouTube...\n", dummyPlaylist.Name)
	playlistURL, _, err := adapter.SavePlaylist(ctx, dummyPlaylist, token.AccessToken, "", func(converted, failed int) {
		fmt.Printf("   -> Progress: %d converted, %d failed\n", converted, failed)
	})
	if err != nil {
		log.Fatalf("Failed to save playlist to YouTube: %v", err)
	}

	fmt.Printf("\nSUCCESS! Playlist created.\n")
	fmt.Printf("View it here: %s\n", playlistURL)
	os.Exit(0)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	code := r.FormValue("code")
	tok, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatalf("Error exchanging code for token: %v", err)
	}

	fmt.Fprintf(w, "Authorization successful! You can close this window and return to the terminal.")
	ch <- tok
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Printf("Could not open browser automatically. Please copy the URL above manually.\n")
	}
}
