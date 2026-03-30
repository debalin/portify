package mock

import (
	"context"
	"fmt"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
)

// --- MockSource ---

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

// --- MockDestination ---

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

func (d *MockDestination) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	return []*converterv1.CanonicalPlaylist{
		{Id: "mock-dest-playlist-123", Name: "My Existing YT Playlist", Description: "A mocked destination"},
	}, nil
}

func (d *MockDestination) CreatePlaylist(ctx context.Context, name string, description string, authToken string) (string, error) {
	return "mock-playlist-id", nil
}

func (d *MockDestination) MatchTrack(ctx context.Context, track *converterv1.CanonicalTrack, authToken string) (string, error) {
	return "mock-video-" + track.Title, nil
}

func (d *MockDestination) AddTrackToPlaylist(ctx context.Context, playlistID string, trackID string, authToken string) error {
	return nil
}

func (d *MockDestination) GetPlaylistURL(playlistID string) string {
	return "https://youtube.com/playlist?list=" + playlistID
}

// --- MockSourceWithTracks ---

type MockSourceWithTracks struct{}

func (s *MockSourceWithTracks) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "spotify",
		Name:        "Spotify (Mock With Tracks)",
		AuthURLHint: s.GetAuthURL(),
	}
}

func (s *MockSourceWithTracks) GetAuthURL() string { return "http://localhost:5175/?code=mock" }

func (s *MockSourceWithTracks) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	return "mock-spotify-token", nil
}

func (s *MockSourceWithTracks) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	return []*converterv1.CanonicalPlaylist{
		{Id: "playlist-with-tracks", Name: "Test Playlist"},
	}, nil
}

func (s *MockSourceWithTracks) FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error) {
	return &converterv1.CanonicalPlaylist{
		Id:   playlistID,
		Name: "Test Playlist",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Bohemian Rhapsody", Artist: "Queen"},
			{Title: "Stairway to Heaven", Artist: "Led Zeppelin"},
			{Title: "Hotel California", Artist: "Eagles"},
		},
	}, nil
}

// --- MockFailingSource ---

type MockFailingSource struct{}

func (s *MockFailingSource) Info() domain.ProviderInfo {
	return domain.ProviderInfo{ID: "failing-source", Name: "Failing Source"}
}
func (s *MockFailingSource) GetAuthURL() string { return "http://fail" }
func (s *MockFailingSource) ExchangeAuthCode(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("auth exchange failed")
}
func (s *MockFailingSource) ListPlaylists(_ context.Context, _ string) ([]*converterv1.CanonicalPlaylist, error) {
	return nil, fmt.Errorf("list playlists failed")
}
func (s *MockFailingSource) FetchPlaylist(_ context.Context, _ string, _ string) (*converterv1.CanonicalPlaylist, error) {
	return nil, fmt.Errorf("fetch playlist failed")
}

// --- MockFailingDestination ---

type MockFailingDestination struct{}

func (d *MockFailingDestination) Info() domain.ProviderInfo {
	return domain.ProviderInfo{ID: "failing-dest", Name: "Failing Dest"}
}
func (d *MockFailingDestination) GetAuthURL() string { return "http://fail" }
func (d *MockFailingDestination) ExchangeAuthCode(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("auth exchange failed")
}
func (d *MockFailingDestination) ListPlaylists(_ context.Context, _ string) ([]*converterv1.CanonicalPlaylist, error) {
	return nil, fmt.Errorf("list playlists failed")
}
func (d *MockFailingDestination) CreatePlaylist(_ context.Context, _ string, _ string, _ string) (string, error) {
	return "", fmt.Errorf("create playlist failed")
}
func (d *MockFailingDestination) MatchTrack(_ context.Context, _ *converterv1.CanonicalTrack, _ string) (string, error) {
	return "", fmt.Errorf("match track failed")
}
func (d *MockFailingDestination) AddTrackToPlaylist(_ context.Context, _ string, _ string, _ string) error {
	return fmt.Errorf("add track failed")
}
func (d *MockFailingDestination) GetPlaylistURL(playlistID string) string {
	return "http://fail"
}
