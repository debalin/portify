package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/common"
)

// --- BuildSearchQuery Tests ---

func TestBuildSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		track    *converterv1.CanonicalTrack
		expected string
	}{
		{
			name:     "standard track",
			track:    &converterv1.CanonicalTrack{Title: "Bohemian Rhapsody", Artist: "Queen"},
			expected: "Bohemian Rhapsody Queen official audio",
		},
		{
			name:     "track with special characters",
			track:    &converterv1.CanonicalTrack{Title: "Don't Stop Me Now", Artist: "Queen"},
			expected: "Don't Stop Me Now Queen official audio",
		},
		{
			name:     "empty artist",
			track:    &converterv1.CanonicalTrack{Title: "Some Track", Artist: ""},
			expected: "Some Track  official audio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSearchQuery(tt.track)
			if got != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

// --- Constructor & Info Tests ---

func TestNewAdapter(t *testing.T) {
	a := NewAdapter()
	if a == nil {
		t.Fatal("Expected non-nil adapter")
	}
	if a.HTTPClient != nil {
		t.Error("Expected nil httpClient by default")
	}
	if a.BaseURL != "" {
		t.Error("Expected empty BaseURL by default")
	}
}

func TestNewAdapterWithOptions(t *testing.T) {
	c := &http.Client{}
	a := NewAdapter(common.WithHTTPClient(c), common.WithBaseURL("http://test:8080"))
	if a.HTTPClient != c {
		t.Error("Expected injected HTTP client")
	}
	if a.BaseURL != "http://test:8080" {
		t.Errorf("Expected BaseURL 'http://test:8080', got '%s'", a.BaseURL)
	}
}

func TestInfo(t *testing.T) {
	a := NewAdapter()
	info := a.Info()
	if info.ID != "youtube" {
		t.Errorf("Expected ID 'youtube', got '%s'", info.ID)
	}
	if info.Name != "YouTube Music" {
		t.Errorf("Expected Name 'YouTube Music', got '%s'", info.Name)
	}
}

func TestGetAuthURL(t *testing.T) {
	a := NewAdapter()
	url := a.GetAuthURL()
	if url == "" {
		t.Error("Expected non-empty auth URL")
	}
}

// --- Helper to create a test adapter ---

func newTestAdapter(serverURL string) *Adapter {
	return NewAdapter(
		common.WithHTTPClient(http.DefaultClient),
		common.WithBaseURL(serverURL),
	)
}

// --- ListPlaylists Tests ---

func TestListPlaylists_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/youtube/v3/playlists") {
			resp := map[string]any{
				"kind": "youtube#playlistListResponse",
				"items": []map[string]any{
					{"id": "PLtest1", "snippet": map[string]any{"title": "My YT Playlist", "description": "A test playlist"}},
					{"id": "PLtest2", "snippet": map[string]any{"title": "Another Playlist", "description": "Second one"}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlists, err := a.ListPlaylists(context.Background(), "mock-token")
	if err != nil {
		t.Fatalf("ListPlaylists returned error: %v", err)
	}
	if len(playlists) != 2 {
		t.Fatalf("Expected 2 playlists, got %d", len(playlists))
	}
	if playlists[0].Id != "PLtest1" {
		t.Errorf("Expected ID 'PLtest1', got '%s'", playlists[0].Id)
	}
}

func TestListPlaylists_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"kind": "youtube#playlistListResponse", "items": []any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": 403, "message": "Quota exceeded"}})
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	_, err := a.ListPlaylists(context.Background(), "mock-token")
	if err == nil {
		t.Fatal("Expected error from 403 response")
	}
}

// --- Helper mux for YouTube API simulation ---

func ytMux(
	createdPlaylistID string,
	searchResults map[string]string, // query substring -> videoId
	insertErrors map[string]bool, // videoId -> should error
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/youtube/v3/playlists"):
			json.NewEncoder(w).Encode(map[string]any{"id": createdPlaylistID})

		case r.Method == "GET" && strings.Contains(r.URL.Path, "/youtube/v3/search"):
			q := r.URL.Query().Get("q")
			videoID := ""
			for substr, vid := range searchResults {
				if strings.Contains(q, substr) {
					videoID = vid
					break
				}
			}
			if videoID == "" {
				json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"id": map[string]any{"kind": "youtube#video", "videoId": videoID}, "snippet": map[string]any{"title": q}},
				},
			})

		case r.Method == "POST" && strings.Contains(r.URL.Path, "/youtube/v3/playlistItems"):
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			videoID := ""
			if snippet, ok := body["snippet"].(map[string]any); ok {
				if rid, ok := snippet["resourceId"].(map[string]any); ok {
					videoID, _ = rid["videoId"].(string)
				}
			}
			if insertErrors[videoID] {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": 500, "message": "Insert failed"}})
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"id": "item-" + videoID})

		default:
			http.NotFound(w, r)
		}
	})
}

// --- CreatePlaylist Tests ---

func TestCreatePlaylist_Success(t *testing.T) {
	server := httptest.NewServer(ytMux("PL-new-123", nil, nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	id, err := a.CreatePlaylist(context.Background(), "Test Playlist", "A description", "mock-token")
	if err != nil {
		t.Fatalf("CreatePlaylist returned error: %v", err)
	}
	if id != "PL-new-123" {
		t.Errorf("Expected playlist ID 'PL-new-123', got '%s'", id)
	}
}

func TestCreatePlaylist_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": 403, "message": "Forbidden"}})
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	_, err := a.CreatePlaylist(context.Background(), "Fail", "Desc", "mock-token")
	if err == nil {
		t.Fatal("Expected error when playlist creation fails")
	}
}

// --- MatchTrack Tests ---

func TestMatchTrack_Success(t *testing.T) {
	server := httptest.NewServer(ytMux("", map[string]string{"Bohemian Rhapsody": "dQw4w9WgXcQ"}, nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	track := &converterv1.CanonicalTrack{Title: "Bohemian Rhapsody", Artist: "Queen"}
	videoID, err := a.MatchTrack(context.Background(), track, "mock-token")
	if err != nil {
		t.Fatalf("MatchTrack returned error: %v", err)
	}
	if videoID != "dQw4w9WgXcQ" {
		t.Errorf("Expected videoID 'dQw4w9WgXcQ', got '%s'", videoID)
	}
}

func TestMatchTrack_NoResults(t *testing.T) {
	server := httptest.NewServer(ytMux("", nil, nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	track := &converterv1.CanonicalTrack{Title: "Nonexistent Song", Artist: "Nobody"}
	videoID, err := a.MatchTrack(context.Background(), track, "mock-token")
	if err != nil {
		t.Fatalf("MatchTrack returned error: %v", err)
	}
	if videoID != "" {
		t.Errorf("Expected empty videoID, got '%s'", videoID)
	}
}

func TestMatchTrack_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	track := &converterv1.CanonicalTrack{Title: "Test", Artist: "Test"}
	_, err := a.MatchTrack(context.Background(), track, "mock-token")
	if err == nil {
		t.Fatal("Expected error from 500 response")
	}
}

// --- AddTrackToPlaylist Tests ---

func TestAddTrackToPlaylist_Success(t *testing.T) {
	server := httptest.NewServer(ytMux("", nil, nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	err := a.AddTrackToPlaylist(context.Background(), "PL-test", "vid-123", "mock-token")
	if err != nil {
		t.Fatalf("AddTrackToPlaylist returned error: %v", err)
	}
}

func TestAddTrackToPlaylist_Error(t *testing.T) {
	server := httptest.NewServer(ytMux("", nil, map[string]bool{"vid-fail": true}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	err := a.AddTrackToPlaylist(context.Background(), "PL-test", "vid-fail", "mock-token")
	if err == nil {
		t.Fatal("Expected error when insert fails")
	}
}

// --- GetPlaylistURL Test ---

func TestGetPlaylistURL(t *testing.T) {
	a := NewAdapter()
	url := a.GetPlaylistURL("PL-abc-123")
	expected := "https://music.youtube.com/playlist?list=PL-abc-123"
	if url != expected {
		t.Errorf("Expected '%s', got '%s'", expected, url)
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
