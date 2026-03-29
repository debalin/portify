package youtube

import (
	"context"
	"fmt"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"os"
)

// Adapter implements domain.PlaylistSink for YouTube
type Adapter struct{}

// NewAdapter creates a new YouTube adapter instance
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Info returns basic information about the YouTube provider
func (a *Adapter) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "youtube",
		Name:        "YouTube Music",
		AuthURLHint: a.GetAuthURL(),
	}
}

func getYouTubeOAuthConfig() *oauth2.Config {
	redirectURL := os.Getenv("FRONTEND_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:5175/"
	}

	return &oauth2.Config{
		ClientID:     os.Getenv("YOUTUBE_ID"),
		ClientSecret: os.Getenv("YOUTUBE_SECRET"),
		RedirectURL:  redirectURL,
		Scopes:       []string{youtube.YoutubeScope},
		Endpoint:     google.Endpoint,
	}
}

func (a *Adapter) GetAuthURL() string {
	return getYouTubeOAuthConfig().AuthCodeURL("youtube", oauth2.AccessTypeOffline)
}

func (a *Adapter) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	token, err := getYouTubeOAuthConfig().Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// ListPlaylists fetches the user's existing YouTube playlists.
func (a *Adapter) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	token := &oauth2.Token{
		AccessToken: authToken,
		TokenType:   "Bearer",
	}
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	service, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
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

// SavePlaylist takes a CanonicalPlaylist and creates it on the user's YouTube account.
// It uses TrackMatcher to find the corresponding YouTube Video IDs for each track before adding them.
// Note: This requires an authToken with the "https://www.googleapis.com/auth/youtube" scope.
func (a *Adapter) SavePlaylist(ctx context.Context, playlist *converterv1.CanonicalPlaylist, authToken string, destinationPlaylistID string, onProgress func(converted, failed int)) (string, []*converterv1.CanonicalTrack, error) {
	token := &oauth2.Token{
		AccessToken: authToken,
		TokenType:   "Bearer",
	}
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// Initialize the YouTube API client
	service, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create YouTube client: %w", err)
	}

	playlistID := destinationPlaylistID

	// 1. Create the empty Playlist only if no target was explicitly given
	if playlistID == "" {
		ytPlaylist := &youtube.Playlist{
			Snippet: &youtube.PlaylistSnippet{
				Title:       playlist.Name,
				Description: playlist.Description + "\n\n(Converted via Playlist Converter)",
			},
			Status: &youtube.PlaylistStatus{
				PrivacyStatus: "private", // Always default to private for safety
			},
		}

		call := service.Playlists.Insert([]string{"snippet", "status"}, ytPlaylist)
		createdPlaylist, err := call.Do()
		if err != nil {
			return "", nil, fmt.Errorf("failed to create playlist on YouTube: %w", err)
		}
		playlistID = createdPlaylist.Id
	}

	converted := 0
	failed := 0
	var failedTracks []*converterv1.CanonicalTrack

	// 2. Search for tracks and add them
	for _, track := range playlist.Tracks {
		// Attempt to match the track directly inside this Adapter since the logic is highly YouTube-specific.
		videoID, err := a.matchTrack(service, track)
		if err != nil {
			fmt.Printf("Warning: Failed to match track %s by %s: %v\n", track.Title, track.Artist, err)
			failed++
			failedTracks = append(failedTracks, track)
			if onProgress != nil {
				onProgress(converted, failed)
			}
			continue // Skip tracks we can't find rather than failing the whole playlist
		}

		if videoID == "" {
			fmt.Printf("Warning: Could not find any suitable match for %s by %s\n", track.Title, track.Artist)
			failed++
			failedTracks = append(failedTracks, track)
			if onProgress != nil {
				onProgress(converted, failed)
			}
			continue
		}

		// Add the found video to the created playlist
		playlistItem := &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: playlistID,
				ResourceId: &youtube.ResourceId{
					Kind:    "youtube#video",
					VideoId: videoID,
				},
			},
		}

		insertCall := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
		_, err = insertCall.Do()
		if err != nil {
			fmt.Printf("Warning: Failed to insert video %s into playlist: %v\n", videoID, err)
			failed++
			failedTracks = append(failedTracks, track)
			if onProgress != nil {
				onProgress(converted, failed)
			}
			continue
		}

		converted++
		if onProgress != nil {
			onProgress(converted, failed)
		}
	}

	// Return the URL to the completed playlist
	playlistURL := fmt.Sprintf("https://music.youtube.com/playlist?list=%s", playlistID)
	return playlistURL, failedTracks, nil
}

// matchTrack implements a rudimentary TrackMatcher specifically for the YouTube API context.
func (a *Adapter) matchTrack(service *youtube.Service, track *converterv1.CanonicalTrack) (string, error) {
	// YouTube search is very text-dependent. The best format is usually "Title Artist"
	searchQuery := fmt.Sprintf("%s %s", track.Title, track.Artist)

	// In YouTube Music, songs are technically just videos with an "Official Audio" or specific metadata categorization.
	// Since we are using the generic YouTube v3 API, we search for videos.
	// To improve accuracy for music, we could append "official audio" or "topic"
	searchQuery += " official audio"

	call := service.Search.List([]string{"id", "snippet"}).
		Q(searchQuery).
		Type("video").
		MaxResults(3) // Get top 3 to inspect

	response, err := call.Do()
	if err != nil {
		return "", err
	}

	if len(response.Items) == 0 {
		return "", nil // No match found
	}

	// For a production app, we would inspect the snippet.Title and snippet.ChannelTitle
	// here to find the closest Levenshtein distance match.
	// For this MVP, we will trust Google's search algorithm and return the top match.
	return response.Items[0].Id.VideoId, nil
}
