package spotify

import (
	"context"
	"fmt"
	"log"
	"strings"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/common"
	"github.com/debalin/portify/internal/domain"
	sp "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// Adapter implements domain.PlaylistSource for Spotify
type Adapter struct {
	common.BaseAdapter
}

// NewAdapter creates a new Spotify adapter instance
func NewAdapter(opts ...common.Option) *Adapter {
	a := &Adapter{
		BaseAdapter: common.BaseAdapter{
			OAuthCfg: common.OAuthConfig{
				ProviderID:   "spotify",
				ClientIDEnv:  "SPOTIFY_ID",
				ClientSecEnv: "SPOTIFY_SECRET",
				Scopes:       []string{"playlist-read-private", "playlist-read-collaborative", "playlist-modify-private", "playlist-modify-public"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.spotify.com/authorize",
					TokenURL: "https://accounts.spotify.com/api/token",
				},
			},
		},
	}
	a.ApplyOptions(opts)
	return a
}

// Info returns basic information about the Spotify provider
func (a *Adapter) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "spotify",
		Name:        "Spotify",
		AuthURLHint: a.GetAuthURL(),
	}
}

func (a *Adapter) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	httpClient := a.GetHTTPClient(ctx, authToken)
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
	httpClient := a.GetHTTPClient(ctx, authToken)
	client := sp.New(httpClient)

	spPlaylist, err := client.GetPlaylist(ctx, sp.ID(playlistID))
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist metadata: %w", err)
	}

	canonical := &converterv1.CanonicalPlaylist{
		Name:        spPlaylist.Name,
		Description: spPlaylist.Description,
		Tracks:      make([]*converterv1.CanonicalTrack, 0, spPlaylist.Tracks.Total),
	}

	offset := 0
	limit := 100

	for {
		trackPage, err := client.GetPlaylistItems(ctx, sp.ID(playlistID), sp.Limit(limit), sp.Offset(offset))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist tracks at offset %d: %w", offset, err)
		}

		for _, item := range trackPage.Items {
			if item.Track.Track == nil {
				continue
			}

			track := item.Track.Track

			artistName := ""
			if len(track.Artists) > 0 {
				artistName = track.Artists[0].Name
			}

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

		if len(canonical.Tracks) >= int(trackPage.Total) || len(trackPage.Items) == 0 {
			break
		}

		offset += limit
	}

	return canonical, nil
}

// CreatePlaylist creates a new, empty playlist on Spotify.
// Returns the platform-specific playlist ID.
func (a *Adapter) CreatePlaylist(ctx context.Context, name string, description string, authToken string) (string, error) {
	httpClient := a.GetHTTPClient(ctx, authToken)
	client := sp.New(httpClient)

	user, err := client.CurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	// Spotify playlist name limit is 100 characters
	if len(name) > 100 {
		name = name[:100]
	}
	// Spotify playlist description must not contain newlines or carriage returns
	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.ReplaceAll(description, "\r", " ")
	for strings.Contains(description, "  ") {
		description = strings.ReplaceAll(description, "  ", " ")
	}
	description = strings.TrimSpace(description)
	// Spotify playlist description limit is 300 characters
	if len(description) > 300 {
		description = description[:300]
	}

	log.Printf("[Spotify] Creating playlist: name=%q, description=%q, userID=%q", name, description, user.ID)
	spPlaylist, err := client.CreatePlaylistForUser(ctx, user.ID, name, description, false, false)
	if err != nil {
		log.Printf("[Spotify] CreatePlaylistForUser failed: %v", err)
		return "", fmt.Errorf("failed to create playlist: %w", err)
	}

	return string(spPlaylist.ID), nil
}

// MatchTrack searches Spotify for a track matching the given canonical track.
// Returns the platform-specific track ID, or empty string if no match was found.
func (a *Adapter) MatchTrack(ctx context.Context, track *converterv1.CanonicalTrack, authToken string) (string, error) {
	httpClient := a.GetHTTPClient(ctx, authToken)
	client := sp.New(httpClient)

	query := fmt.Sprintf("track:%s artist:%s", track.Title, track.Artist)
	results, err := client.Search(ctx, query, sp.SearchTypeTrack)
	if err != nil {
		return "", fmt.Errorf("failed to search track: %w", err)
	}

	if results.Tracks != nil && len(results.Tracks.Tracks) > 0 {
		return string(results.Tracks.Tracks[0].ID), nil
	}

	// Fallback: search just by title if title+artist yields nothing
	queryFallback := fmt.Sprintf("track:%s", track.Title)
	resultsFallback, err := client.Search(ctx, queryFallback, sp.SearchTypeTrack)
	if err != nil {
		return "", fmt.Errorf("failed to search track: %w", err)
	}

	if resultsFallback.Tracks != nil && len(resultsFallback.Tracks.Tracks) > 0 {
		return string(resultsFallback.Tracks.Tracks[0].ID), nil
	}

	return "", nil
}

// AddTrackToPlaylist inserts a single matched track into a playlist.
func (a *Adapter) AddTrackToPlaylist(ctx context.Context, playlistID string, trackID string, authToken string) error {
	httpClient := a.GetHTTPClient(ctx, authToken)
	client := sp.New(httpClient)

	_, err := client.AddTracksToPlaylist(ctx, sp.ID(playlistID), sp.ID(trackID))
	if err != nil {
		return fmt.Errorf("failed to insert track %s into playlist: %w", trackID, err)
	}
	return nil
}

// GetPlaylistURL returns the user-facing URL for a playlist given its platform ID.
func (a *Adapter) GetPlaylistURL(playlistID string) string {
	return fmt.Sprintf("https://open.spotify.com/playlist/%s", playlistID)
}
