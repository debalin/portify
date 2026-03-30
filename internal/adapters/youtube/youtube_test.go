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

// --- BuildSearchQuery Tests (pure function, no mocking needed) ---

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
					{
						"id": "PLtest1",
						"snippet": map[string]any{
							"title":       "My YT Playlist",
							"description": "A test playlist",
						},
					},
					{
						"id": "PLtest2",
						"snippet": map[string]any{
							"title":       "Another Playlist",
							"description": "Second one",
						},
					},
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
	if playlists[0].Name != "My YT Playlist" {
		t.Errorf("Expected name 'My YT Playlist', got '%s'", playlists[0].Name)
	}
	if playlists[1].Id != "PLtest2" {
		t.Errorf("Expected ID 'PLtest2', got '%s'", playlists[1].Id)
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
		resp := map[string]any{
			"error": map[string]any{
				"code":    403,
				"message": "Quota exceeded",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	_, err := a.ListPlaylists(context.Background(), "mock-token")
	if err == nil {
		t.Fatal("Expected error from 403 response")
	}
}

// --- SavePlaylist Tests ---

// ytMux builds a test HTTP mux simulating YouTube's API endpoints.
// searchResults controls what matchTrack returns per search query.
// insertErrors can be set to make specific playlist item inserts fail.
func ytMux(
	createdPlaylistID string,
	searchResults map[string]string, // query substring -> videoId
	insertErrors map[string]bool, // videoId -> should error
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// Create playlist
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/youtube/v3/playlists"):
			resp := map[string]any{"id": createdPlaylistID}
			json.NewEncoder(w).Encode(resp)

		// Search for video
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
				// No match
				resp := map[string]any{"items": []any{}}
				json.NewEncoder(w).Encode(resp)
				return
			}
			resp := map[string]any{
				"items": []map[string]any{
					{
						"id": map[string]any{
							"kind":    "youtube#video",
							"videoId": videoID,
						},
						"snippet": map[string]any{"title": q},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)

		// Insert playlist item
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
				resp := map[string]any{"error": map[string]any{"code": 500, "message": "Insert failed"}}
				json.NewEncoder(w).Encode(resp)
				return
			}

			resp := map[string]any{"id": "item-" + videoID}
			json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	})
}

func TestSavePlaylist_CreateNewPlaylist_AllTracksFound(t *testing.T) {
	server := httptest.NewServer(ytMux(
		"PL-new-123",
		map[string]string{
			"Bohemian Rhapsody": "vid-queen-1",
			"Hotel California":  "vid-eagles-1",
		},
		nil, // no insert errors
	))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name:        "Test Playlist",
		Description: "A test",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Bohemian Rhapsody", Artist: "Queen"},
			{Title: "Hotel California", Artist: "Eagles"},
		},
	}

	var progressCalls []struct{ converted, failed int }
	url, failedTracks, err := a.SavePlaylist(
		context.Background(), playlist, "mock-token", "",
		func(converted, failed int) {
			progressCalls = append(progressCalls, struct{ converted, failed int }{converted, failed})
		},
	)

	if err != nil {
		t.Fatalf("SavePlaylist returned error: %v", err)
	}
	if !strings.Contains(url, "PL-new-123") {
		t.Errorf("Expected URL containing 'PL-new-123', got '%s'", url)
	}
	if len(failedTracks) != 0 {
		t.Errorf("Expected 0 failed tracks, got %d", len(failedTracks))
	}
	if len(progressCalls) != 2 {
		t.Errorf("Expected 2 progress calls, got %d", len(progressCalls))
	}
	// Last progress: 2 converted, 0 failed
	last := progressCalls[len(progressCalls)-1]
	if last.converted != 2 || last.failed != 0 {
		t.Errorf("Expected final progress (2,0), got (%d,%d)", last.converted, last.failed)
	}
}

func TestSavePlaylist_AppendToExisting(t *testing.T) {
	server := httptest.NewServer(ytMux(
		"", // shouldn't be used since we provide destinationPlaylistID
		map[string]string{
			"Bohemian Rhapsody": "vid-queen-1",
		},
		nil,
	))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name: "Test",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Bohemian Rhapsody", Artist: "Queen"},
		},
	}

	url, _, err := a.SavePlaylist(
		context.Background(), playlist, "mock-token", "EXISTING-PL-456", nil,
	)

	if err != nil {
		t.Fatalf("SavePlaylist returned error: %v", err)
	}
	if !strings.Contains(url, "EXISTING-PL-456") {
		t.Errorf("Expected URL containing 'EXISTING-PL-456', got '%s'", url)
	}
}

func TestSavePlaylist_SomeTracksNotFound(t *testing.T) {
	server := httptest.NewServer(ytMux(
		"PL-new-789",
		map[string]string{
			"Bohemian Rhapsody": "vid-queen-1",
			// "Nonexistent Song" is NOT in search results
		},
		nil,
	))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name: "Mixed Results",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Bohemian Rhapsody", Artist: "Queen"},
			{Title: "Nonexistent Song", Artist: "Nobody"},
		},
	}

	var progressCalls []struct{ converted, failed int }
	_, failedTracks, err := a.SavePlaylist(
		context.Background(), playlist, "mock-token", "",
		func(converted, failed int) {
			progressCalls = append(progressCalls, struct{ converted, failed int }{converted, failed})
		},
	)

	if err != nil {
		t.Fatalf("SavePlaylist returned error: %v", err)
	}
	if len(failedTracks) != 1 {
		t.Fatalf("Expected 1 failed track, got %d", len(failedTracks))
	}
	if failedTracks[0].Title != "Nonexistent Song" {
		t.Errorf("Expected failed track 'Nonexistent Song', got '%s'", failedTracks[0].Title)
	}
	// Final progress should show 1 converted, 1 failed
	last := progressCalls[len(progressCalls)-1]
	if last.converted != 1 || last.failed != 1 {
		t.Errorf("Expected final progress (1,1), got (%d,%d)", last.converted, last.failed)
	}
}

func TestSavePlaylist_InsertError(t *testing.T) {
	server := httptest.NewServer(ytMux(
		"PL-insert-err",
		map[string]string{
			"Bohemian Rhapsody": "vid-queen-1",
			"Hotel California":  "vid-eagles-1",
		},
		map[string]bool{
			"vid-eagles-1": true, // This insert will fail
		},
	))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name: "Insert Error Test",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Bohemian Rhapsody", Artist: "Queen"},
			{Title: "Hotel California", Artist: "Eagles"},
		},
	}

	var progressCalls []struct{ converted, failed int }
	_, failedTracks, err := a.SavePlaylist(
		context.Background(), playlist, "mock-token", "",
		func(converted, failed int) {
			progressCalls = append(progressCalls, struct{ converted, failed int }{converted, failed})
		},
	)

	if err != nil {
		t.Fatalf("SavePlaylist returned error: %v", err)
	}
	if len(failedTracks) != 1 {
		t.Fatalf("Expected 1 failed track, got %d", len(failedTracks))
	}
	if failedTracks[0].Title != "Hotel California" {
		t.Errorf("Expected failed track 'Hotel California', got '%s'", failedTracks[0].Title)
	}
}

func TestSavePlaylist_CreatePlaylistError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]any{"error": map[string]any{"code": 403, "message": "Forbidden"}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name:   "Should Fail",
		Tracks: []*converterv1.CanonicalTrack{{Title: "Test", Artist: "Test"}},
	}

	_, _, err := a.SavePlaylist(context.Background(), playlist, "mock-token", "", nil)
	if err == nil {
		t.Fatal("Expected error when playlist creation fails")
	}
}

func TestSavePlaylist_NilProgressCallback(t *testing.T) {
	server := httptest.NewServer(ytMux(
		"PL-nil-cb",
		map[string]string{"Test": "vid-1"},
		nil,
	))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name:   "No Callback",
		Tracks: []*converterv1.CanonicalTrack{{Title: "Test", Artist: "Artist"}},
	}

	// Should not panic with nil callback
	url, _, err := a.SavePlaylist(context.Background(), playlist, "mock-token", "", nil)
	if err != nil {
		t.Fatalf("SavePlaylist returned error: %v", err)
	}
	if !strings.Contains(url, "PL-nil-cb") {
		t.Errorf("Expected URL containing 'PL-nil-cb', got '%s'", url)
	}
}

func TestSavePlaylist_EmptyPlaylist(t *testing.T) {
	server := httptest.NewServer(ytMux("PL-empty", nil, nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist := &converterv1.CanonicalPlaylist{
		Name:   "Empty",
		Tracks: []*converterv1.CanonicalTrack{},
	}

	url, failedTracks, err := a.SavePlaylist(context.Background(), playlist, "mock-token", "", nil)
	if err != nil {
		t.Fatalf("SavePlaylist returned error: %v", err)
	}
	if !strings.Contains(url, "PL-empty") {
		t.Errorf("Expected URL containing 'PL-empty', got '%s'", url)
	}
	if len(failedTracks) != 0 {
		t.Errorf("Expected 0 failed tracks, got %d", len(failedTracks))
	}
}

// --- matchTrack Tests (through the adapter) ---

func TestMatchTrack_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id": map[string]any{
						"kind":    "youtube#video",
						"videoId": "dQw4w9WgXcQ",
					},
					"snippet": map[string]any{"title": "Test Video"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	service, err := a.newService(context.Background(), "mock-token")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	track := &converterv1.CanonicalTrack{Title: "Test", Artist: "Artist"}
	videoID, err := a.matchTrack(service, track)
	if err != nil {
		t.Fatalf("matchTrack returned error: %v", err)
	}
	if videoID != "dQw4w9WgXcQ" {
		t.Errorf("Expected videoID 'dQw4w9WgXcQ', got '%s'", videoID)
	}
}

func TestMatchTrack_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"items": []any{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	service, _ := a.newService(context.Background(), "mock-token")

	track := &converterv1.CanonicalTrack{Title: "Nonexistent Song", Artist: "Nobody"}
	videoID, err := a.matchTrack(service, track)
	if err != nil {
		t.Fatalf("matchTrack returned error: %v", err)
	}
	if videoID != "" {
		t.Errorf("Expected empty videoID for no results, got '%s'", videoID)
	}
}

func TestMatchTrack_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	a := newTestAdapter(server.URL)
	service, _ := a.newService(context.Background(), "mock-token")

	track := &converterv1.CanonicalTrack{Title: "Test", Artist: "Test"}
	_, err := a.matchTrack(service, track)
	if err == nil {
		t.Fatal("Expected error from 500 response")
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
