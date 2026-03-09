package spotify

import (
	"context"
	"fmt"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"os"
)

// Adapter implements domain.PlaylistSource for Spotify
type Adapter struct{}

// NewAdapter creates a new Spotify adapter instance
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Info returns basic information about the Spotify provider
func (a *Adapter) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "spotify",
		Name:        "Spotify",
		AuthURLHint: a.GetAuthURL(),
	}
}

func getSpotifyOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		RedirectURL:  "http://localhost:5175/",
		Scopes:       []string{"playlist-read-private", "playlist-read-collaborative", "playlist-modify-private", "playlist-modify-public"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}
}

func (a *Adapter) GetAuthURL() string {
	return getSpotifyOAuthConfig().AuthCodeURL("spotify", oauth2.AccessTypeOffline)
}

func (a *Adapter) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	token, err := getSpotifyOAuthConfig().Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (a *Adapter) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	token := &oauth2.Token{
		AccessToken: authToken,
		TokenType:   "Bearer",
	}
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	page, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user playlists: %w", err)
	}

	var playlists []*converterv1.CanonicalPlaylist
	for _, sp := range page.Playlists {
		playlists = append(playlists, &converterv1.CanonicalPlaylist{
			Id:          string(sp.ID),
			Name:        sp.Name,
			Description: sp.Description,
		})
	}
	return playlists, nil
}

// FetchPlaylist fetches a complete playlist from Spotify and maps it to the generic CanonicalPlaylist
func (a *Adapter) FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error) {
	// Create an authenticated Spotify client using the provided user access token
	token := &oauth2.Token{
		AccessToken: authToken,
		TokenType:   "Bearer",
	}
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	// Fetch basic playlist details (name, description, etc.)
	spPlaylist, err := client.GetPlaylist(ctx, spotify.ID(playlistID))
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist metadata: %w", err)
	}

	canonical := &converterv1.CanonicalPlaylist{
		Name:        spPlaylist.Name,
		Description: spPlaylist.Description,
		Tracks:      make([]*converterv1.CanonicalTrack, 0, spPlaylist.Tracks.Total),
	}

	// Fetch all tracks with pagination
	offset := 0
	limit := 100 // Maximum allowed by Spotify API

	for {
		trackPage, err := client.GetPlaylistItems(ctx, spotify.ID(playlistID), spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist tracks at offset %d: %w", offset, err)
		}

		for _, item := range trackPage.Items {
			// Skip items if there isn't actually a track (e.g. episodic content or local files without a valid track attached)
			if item.Track.Track == nil {
				continue
			}

			track := item.Track.Track

			// Determine primary artist
			artistName := ""
			if len(track.Artists) > 0 {
				artistName = track.Artists[0].Name
			}

			// Extract ISRC if available (mostly useful for track matching)
			isrc := ""
			if val, ok := track.ExternalIDs["isrc"]; ok {
				isrc = val
			}

			canonicalTrack := &converterv1.CanonicalTrack{
				Title:      track.Name,
				Artist:     artistName,
				Album:      track.Album.Name,
				DurationMs: int64(track.Duration),
				Isrc:       isrc,
			}

			canonical.Tracks = append(canonical.Tracks, canonicalTrack)
		}

		// If we fetched less than limit or there's no next page, we are done
		if len(trackPage.Items) < limit || trackPage.Next == "" {
			break
		}

		offset += limit
	}

	return canonical, nil
}
