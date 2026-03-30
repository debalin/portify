package spotify

import (
	"context"
	"fmt"
	"net/http"
	"os"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
	sp "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// Adapter implements domain.PlaylistSource for Spotify
type Adapter struct {
	httpClient *http.Client // If set, used instead of creating from token (for testing)
}

// Option configures an Adapter.
type Option func(*Adapter)

// WithHTTPClient injects a custom HTTP client (used for testing).
func WithHTTPClient(c *http.Client) Option {
	return func(a *Adapter) {
		a.httpClient = c
	}
}

// NewAdapter creates a new Spotify adapter instance
func NewAdapter(opts ...Option) *Adapter {
	a := &Adapter{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// getClient returns the injected HTTP client or creates one from the auth token.
func (a *Adapter) getClient(ctx context.Context, authToken string) *http.Client {
	if a.httpClient != nil {
		return a.httpClient
	}
	token := &oauth2.Token{
		AccessToken: authToken,
		TokenType:   "Bearer",
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
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
	redirectURL := os.Getenv("FRONTEND_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:5175/"
	}

	return &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		RedirectURL:  redirectURL,
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
	httpClient := a.getClient(ctx, authToken)
	client := sp.New(httpClient)

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
	httpClient := a.getClient(ctx, authToken)
	client := sp.New(httpClient)

	// Fetch basic playlist details (name, description, etc.)
	spPlaylist, err := client.GetPlaylist(ctx, sp.ID(playlistID))
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
		trackPage, err := client.GetPlaylistItems(ctx, sp.ID(playlistID), sp.Limit(limit), sp.Offset(offset))
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

		// If we fetched the total number of available tracks, break the pagination loop
		if len(canonical.Tracks) >= int(trackPage.Total) || len(trackPage.Items) == 0 {
			break
		}

		offset += limit
	}

	return canonical, nil
}
