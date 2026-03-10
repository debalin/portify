package mock

import (
	"context"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
)

type MockSource struct{}

func (s *MockSource) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "spotify",
		Name:        "Spotify (Mock)",
		AuthURLHint: s.GetAuthURL(),
	}
}

func (s *MockSource) GetAuthURL() string {
	return "http://localhost:5175/?code=mock-auth-code"
}

func (s *MockSource) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	return "mock-spotify-token", nil
}

func (s *MockSource) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	return []*converterv1.CanonicalPlaylist{
		{Id: "mock-playlist-123", Name: "Mock Favorites", Description: "A mocked playlist from Spotify"},
		{Id: "mock-playlist-456", Name: "Coding Focus", Description: "Mocked focus beats"},
	}, nil
}

func (s *MockSource) FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error) {
	return &converterv1.CanonicalPlaylist{}, nil
}

type MockDestination struct{}

func (d *MockDestination) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "youtube",
		Name:        "YouTube Music (Mock)",
		AuthURLHint: d.GetAuthURL(),
	}
}

func (d *MockDestination) GetAuthURL() string {
	return "http://localhost:5175/?code=mock-auth-code"
}

func (d *MockDestination) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	return "mock-youtube-token", nil
}

func (d *MockDestination) SavePlaylist(ctx context.Context, playlist *converterv1.CanonicalPlaylist, authToken string, destinationPlaylistID string, onProgress func(converted, failed int)) (string, error) {
	if onProgress != nil {
		for i := 1; i <= len(playlist.Tracks); i++ {
			onProgress(i, 0)
		}
	}
	return "https://youtube.com/playlist?list=mock", nil
}
