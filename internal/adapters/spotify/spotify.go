package spotify

import (
	"context"
	"fmt"

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
