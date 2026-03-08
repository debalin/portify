package mock

import (
	"context"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
)

type MockSource struct{}

func (s *MockSource) Info() domain.ProviderInfo {
	return domain.ProviderInfo{ID: "spotify", Name: "Spotify (Mock)"}
}

func (s *MockSource) FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error) {
	return &converterv1.CanonicalPlaylist{}, nil
}

type MockDestination struct{}

func (d *MockDestination) Info() domain.ProviderInfo {
	return domain.ProviderInfo{ID: "youtube", Name: "YouTube Music (Mock)"}
}

func (d *MockDestination) SavePlaylist(ctx context.Context, playlist *converterv1.CanonicalPlaylist, authToken string) (string, error) {
	return "https://youtube.com/playlist?list=mock", nil
}
