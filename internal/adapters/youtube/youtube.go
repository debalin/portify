package youtube

import (
	"context"
	"fmt"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/common"
	"github.com/debalin/portify/internal/domain"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

// Adapter implements domain.PlaylistSink for YouTube
type Adapter struct {
	common.BaseAdapter
}

// NewAdapter creates a new YouTube adapter instance
func NewAdapter(opts ...common.Option) *Adapter {
	a := &Adapter{
		BaseAdapter: common.BaseAdapter{
			OAuthCfg: common.OAuthConfig{
				ProviderID:   "youtube",
				ClientIDEnv:  "YOUTUBE_ID",
				ClientSecEnv: "YOUTUBE_SECRET",
				Scopes:       []string{yt.YoutubeScope},
				Endpoint:     google.Endpoint,
			},
		},
	}
	a.ApplyOptions(opts)
	return a
}

// newService creates a YouTube API service, using the injected client if available.
func (a *Adapter) newService(ctx context.Context, authToken string) (*yt.Service, error) {
	httpClient := a.GetHTTPClient(ctx, authToken)
	service, err := yt.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	if a.BaseURL != "" {
		service.BasePath = a.BaseURL
	}
	return service, nil
}

// Info returns basic information about the YouTube provider
func (a *Adapter) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "youtube",
		Name:        "YouTube Music",
		AuthURLHint: a.GetAuthURL(),
	}
}

// ListPlaylists fetches the user's existing YouTube playlists.
func (a *Adapter) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	service, err := a.newService(ctx, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube client: %w", err)
	}

	call := service.Playlists.List([]string{"snippet"}).Mine(true).MaxResults(50)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list youtube playlists: %w", err)
	}

	var canonicals []*converterv1.CanonicalPlaylist
	for _, item := range response.Items {
		canonicals = append(canonicals, &converterv1.CanonicalPlaylist{
			Id:          item.Id,
			Name:        item.Snippet.Title,
			Description: item.Snippet.Description,
		})
	}

	return canonicals, nil
}

// CreatePlaylist creates a new, empty playlist on YouTube.
// Returns the platform-specific playlist ID.
func (a *Adapter) CreatePlaylist(ctx context.Context, name string, description string, authToken string) (string, error) {
	service, err := a.newService(ctx, authToken)
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube client: %w", err)
	}

	ytPlaylist := &yt.Playlist{
		Snippet: &yt.PlaylistSnippet{
			Title:       name,
			Description: description + "\n\n(Converted via Portify)",
		},
		Status: &yt.PlaylistStatus{
			PrivacyStatus: "private",
		},
	}

	call := service.Playlists.Insert([]string{"snippet", "status"}, ytPlaylist)
	created, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("failed to create playlist on YouTube: %w", err)
	}

	return created.Id, nil
}

// MatchTrack searches YouTube for a video matching the given canonical track.
// Returns the YouTube video ID, or empty string if no match was found.
func (a *Adapter) MatchTrack(ctx context.Context, track *converterv1.CanonicalTrack, authToken string) (string, error) {
	service, err := a.newService(ctx, authToken)
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube client: %w", err)
	}

	searchQuery := BuildSearchQuery(track)

	call := service.Search.List([]string{"id", "snippet"}).
		Q(searchQuery).
		Type("video").
		MaxResults(3)

	response, err := call.Do()
	if err != nil {
		return "", err
	}

	if len(response.Items) == 0 {
		return "", nil
	}

	return response.Items[0].Id.VideoId, nil
}

// AddTrackToPlaylist inserts a single matched video into a YouTube playlist.
func (a *Adapter) AddTrackToPlaylist(ctx context.Context, playlistID string, trackID string, authToken string) error {
	service, err := a.newService(ctx, authToken)
	if err != nil {
		return fmt.Errorf("failed to create YouTube client: %w", err)
	}

	playlistItem := &yt.PlaylistItem{
		Snippet: &yt.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &yt.ResourceId{
				Kind:    "youtube#video",
				VideoId: trackID,
			},
		},
	}

	insertCall := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err = insertCall.Do()
	if err != nil {
		return fmt.Errorf("failed to insert video %s into playlist: %w", trackID, err)
	}

	return nil
}

// GetPlaylistURL returns the YouTube Music URL for a playlist.
func (a *Adapter) GetPlaylistURL(playlistID string) string {
	return fmt.Sprintf("https://music.youtube.com/playlist?list=%s", playlistID)
}

// BuildSearchQuery constructs the YouTube search query for a given track.
func BuildSearchQuery(track *converterv1.CanonicalTrack) string {
	return fmt.Sprintf("%s %s official audio", track.Title, track.Artist)
}
