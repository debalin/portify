package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/debalin/portify/internal/adapters/spotify"
	"github.com/joho/godotenv"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	// Load the .env file if it exists
	if err := godotenv.Load("../../.env"); err != nil {
		// Also try looking in current directory depending on where go run is executed from
		_ = godotenv.Load(".env")
	}

	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("You must set SPOTIFY_ID and SPOTIFY_SECRET environment variables. Get them from https://developer.spotify.com/dashboard")
	}

	ctx := context.Background()

	// Use Client Credentials flow for server-to-server public data access
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("Could not get client credentials token: %v", err)
	}

	fmt.Println("Authorization successful using Client Credentials! Token received.")

	// Test with a known public playlist
	playlistID := "3cEYpjA9oz9GiPac4AsH4n"

	adapter := spotify.NewAdapter()

	fmt.Printf("\nFetching playlist %s...\n", playlistID)
	// We're using the client credentials token here
	playlist, err := adapter.FetchPlaylist(ctx, playlistID, token.AccessToken)
	if err != nil {
		log.Fatalf("Failed to fetch playlist: %v", err)
	}

	fmt.Printf("\n=== Playlist Information ===\n")
	fmt.Printf("Name: %s\n", playlist.Name)
	fmt.Printf("Description: %s\n", playlist.Description)
	fmt.Printf("Total Tracks: %d\n\n", len(playlist.Tracks))

	fmt.Printf("=== First 5 Tracks ===\n")
	limit := 5
	if len(playlist.Tracks) < 5 {
		limit = len(playlist.Tracks)
	}

	for i := 0; i < limit; i++ {
		track := playlist.Tracks[i]
		fmt.Printf("%d. %s - %s (Album: %s) [%s]\n", i+1, track.Title, track.Artist, track.Album, track.Isrc)
	}
}
