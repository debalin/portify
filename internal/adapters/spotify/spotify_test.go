package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/debalin/portify/internal/adapters/common"
)

// rewriteTransport redirects all HTTP requests to a test server.
type rewriteTransport struct {
	targetURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.targetURL
	return http.DefaultTransport.RoundTrip(req)
}

func testClient(serverURL string) *http.Client {
	// Strip the "http://" prefix to get just host:port
	host := serverURL[len("http://"):]
	return &http.Client{
		Transport: &rewriteTransport{targetURL: host},
	}
}

// --- Info & Constructor Tests ---

func TestNewAdapter(t *testing.T) {
	a := NewAdapter()
	if a == nil {
		t.Fatal("Expected non-nil adapter")
	}
	if a.HTTPClient != nil {
		t.Error("Expected nil HTTPClient by default")
	}
}

func TestNewAdapterWithHTTPClient(t *testing.T) {
	c := &http.Client{}
	a := NewAdapter(common.WithHTTPClient(c))
	if a.HTTPClient != c {
		t.Error("Expected injected HTTP client")
	}
}

func TestInfo(t *testing.T) {
	a := NewAdapter()
	info := a.Info()
	if info.ID != "spotify" {
		t.Errorf("Expected ID 'spotify', got '%s'", info.ID)
	}
	if info.Name != "Spotify" {
		t.Errorf("Expected Name 'Spotify', got '%s'", info.Name)
	}
}

func TestGetAuthURL(t *testing.T) {
	a := NewAdapter()
	url := a.GetAuthURL()
	if url == "" {
		t.Error("Expected non-empty auth URL")
	}
}

// --- ListPlaylists Tests ---

func TestListPlaylists_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/me/playlists" {
			resp := map[string]any{
				"items": []map[string]any{
					{"id": "pl-1", "name": "My Playlist", "description": "First playlist"},
					{"id": "pl-2", "name": "Coding Beats", "description": "Focus music"},
				},
				"total": 2,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	a := NewAdapter(common.WithHTTPClient(testClient(server.URL)))
	playlists, err := a.ListPlaylists(context.Background(), "mock-token")
	if err != nil {
		t.Fatalf("ListPlaylists returned error: %v", err)
	}
	if len(playlists) != 2 {
		t.Fatalf("Expected 2 playlists, got %d", len(playlists))
	}
	if playlists[0].Id != "pl-1" {
		t.Errorf("Expected ID 'pl-1', got '%s'", playlists[0].Id)
	}
	if playlists[0].Name != "My Playlist" {
		t.Errorf("Expected name 'My Playlist', got '%s'", playlists[0].Name)
	}
	if playlists[1].Id != "pl-2" {
		t.Errorf("Expected ID 'pl-2', got '%s'", playlists[1].Id)
	}
}

func TestListPlaylists_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"items": []any{}, "total": 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := NewAdapter(common.WithHTTPClient(testClient(server.URL)))
	playlists, err := a.ListPlaylists(context.Background(), "mock-token")
	if err != nil {
		t.Fatalf("ListPlaylists returned error: %v", err)
	}
	if len(playlists) != 0 {
		t.Errorf("Expected 0 playlists, got %d", len(playlists))
	}
}

func TestListPlaylists_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"status": 500, "message": "Internal error"}}`))
	}))
	defer server.Close()

	a := NewAdapter(common.WithHTTPClient(testClient(server.URL)))
	_, err := a.ListPlaylists(context.Background(), "mock-token")
	if err == nil {
		t.Fatal("Expected error from 500 response")
	}
}

// --- FetchPlaylist Tests ---

func TestFetchPlaylist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v1/playlists/test-playlist-id":
			// GetPlaylist response
			resp := map[string]any{
				"id":          "test-playlist-id",
				"name":        "My Rock Playlist",
				"description": "Classic rock hits",
				"tracks": map[string]any{
					"total": 2,
				},
			}
			json.NewEncoder(w).Encode(resp)

		case "/v1/playlists/test-playlist-id/tracks":
			// GetPlaylistItems response
			resp := map[string]any{
				"total": 2,
				"items": []map[string]any{
					{
						"track": map[string]any{
							"type":         "track",
							"name":         "Bohemian Rhapsody",
							"duration_ms":  354000,
							"artists":      []map[string]any{{"name": "Queen"}},
							"album":        map[string]any{"name": "A Night at the Opera"},
							"external_ids": map[string]string{"isrc": "GBUM71029604"},
						},
					},
					{
						"track": map[string]any{
							"type":         "track",
							"name":         "Hotel California",
							"duration_ms":  391000,
							"artists":      []map[string]any{{"name": "Eagles"}},
							"album":        map[string]any{"name": "Hotel California"},
							"external_ids": map[string]string{"isrc": "USEE19900001"},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	a := NewAdapter(common.WithHTTPClient(testClient(server.URL)))
	playlist, err := a.FetchPlaylist(context.Background(), "test-playlist-id", "mock-token")
	if err != nil {
		t.Fatalf("FetchPlaylist returned error: %v", err)
	}

	if playlist.Name != "My Rock Playlist" {
		t.Errorf("Expected name 'My Rock Playlist', got '%s'", playlist.Name)
	}
	if playlist.Description != "Classic rock hits" {
		t.Errorf("Expected description 'Classic rock hits', got '%s'", playlist.Description)
	}
	if len(playlist.Tracks) != 2 {
		t.Fatalf("Expected 2 tracks, got %d", len(playlist.Tracks))
	}

	track1 := playlist.Tracks[0]
	if track1.Title != "Bohemian Rhapsody" {
		t.Errorf("Expected title 'Bohemian Rhapsody', got '%s'", track1.Title)
	}
	if track1.Artist != "Queen" {
		t.Errorf("Expected artist 'Queen', got '%s'", track1.Artist)
	}
	if track1.Album != "A Night at the Opera" {
		t.Errorf("Expected album 'A Night at the Opera', got '%s'", track1.Album)
	}
	if track1.Isrc != "GBUM71029604" {
		t.Errorf("Expected ISRC 'GBUM71029604', got '%s'", track1.Isrc)
	}
}

func TestFetchPlaylist_MetadataError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": {"status": 404, "message": "Not found"}}`))
	}))
	defer server.Close()

	a := NewAdapter(common.WithHTTPClient(testClient(server.URL)))
	_, err := a.FetchPlaylist(context.Background(), "bad-id", "mock-token")
	if err == nil {
		t.Fatal("Expected error for nonexistent playlist")
	}
}

func TestFetchPlaylist_TrackWithNoArtist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v1/playlists/no-artist-playlist":
			resp := map[string]any{
				"id": "no-artist-playlist", "name": "Test", "description": "",
				"tracks": map[string]any{"total": 1},
			}
			json.NewEncoder(w).Encode(resp)

		case "/v1/playlists/no-artist-playlist/tracks":
			resp := map[string]any{
				"total": 1,
				"items": []map[string]any{
					{
						"track": map[string]any{
							"type":         "track",
							"name":         "Unknown Track",
							"duration_ms":  180000,
							"artists":      []map[string]any{}, // No artists
							"album":        map[string]any{"name": "Unknown Album"},
							"external_ids": map[string]string{},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	a := NewAdapter(common.WithHTTPClient(testClient(server.URL)))
	playlist, err := a.FetchPlaylist(context.Background(), "no-artist-playlist", "mock-token")
	if err != nil {
		t.Fatalf("FetchPlaylist returned error: %v", err)
	}
	if len(playlist.Tracks) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(playlist.Tracks))
	}
	if playlist.Tracks[0].Artist != "" {
		t.Errorf("Expected empty artist, got '%s'", playlist.Tracks[0].Artist)
	}
	if playlist.Tracks[0].Isrc != "" {
		t.Errorf("Expected empty ISRC, got '%s'", playlist.Tracks[0].Isrc)
	}
}

// --- getClient Tests ---

func TestGetClient_WithInjected(t *testing.T) {
	injected := &http.Client{}
	a := NewAdapter(common.WithHTTPClient(injected))
	got := a.GetHTTPClient(context.Background(), "any-token")
	if got != injected {
		t.Error("Expected injected client to be returned")
	}
}

func TestGetClient_WithoutInjected(t *testing.T) {
	a := NewAdapter()
	got := a.GetHTTPClient(context.Background(), "test-token")
	if got == nil {
		t.Error("Expected non-nil client")
	}
}
