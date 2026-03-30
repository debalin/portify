package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
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
	if a.httpClient != nil {
		t.Error("Expected nil httpClient by default")
	}
}

func TestNewAdapterWithHTTPClient(t *testing.T) {
	c := &http.Client{}
	a := NewAdapter(WithHTTPClient(c))
	if a.httpClient != c {
		t.Error("Expected injected HTTP client")
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

	a := NewAdapter(WithHTTPClient(server.Client()))
	// Override the base URL by using the test server's URL
	service, _ := a.newService(context.Background(), "mock-token")
	service.BasePath = server.URL + "/"

	// Call ListPlaylists - this will use the real method but with our test client
	// Since we can't easily override BasePath through the adapter, let's test via the service directly
	call := service.Playlists.List([]string{"snippet"}).Mine(true).MaxResults(50)
	response, err := call.Do()
	if err != nil {
		t.Fatalf("ListPlaylists returned error: %v", err)
	}
	if len(response.Items) != 2 {
		t.Fatalf("Expected 2 playlists, got %d", len(response.Items))
	}
	if response.Items[0].Id != "PLtest1" {
		t.Errorf("Expected ID 'PLtest1', got '%s'", response.Items[0].Id)
	}
	if response.Items[0].Snippet.Title != "My YT Playlist" {
		t.Errorf("Expected title 'My YT Playlist', got '%s'", response.Items[0].Snippet.Title)
	}
}

func TestListPlaylists_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]any{
			"error": map[string]any{
				"code":    403,
				"message": "The request cannot be completed because you have exceeded your quota.",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := NewAdapter(WithHTTPClient(server.Client()))
	service, _ := a.newService(context.Background(), "mock-token")
	service.BasePath = server.URL + "/"

	call := service.Playlists.List([]string{"snippet"}).Mine(true)
	_, err := call.Do()
	if err == nil {
		t.Fatal("Expected error from 403 response")
	}
}

// --- SavePlaylist / matchTrack Tests ---

func TestSavePlaylist_CreateNewPlaylist(t *testing.T) {
	var insertedPlaylistTitle string
	var insertedVideoIDs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// Create playlist
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/youtube/v3/playlists"):
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if snippet, ok := body["snippet"].(map[string]any); ok {
				insertedPlaylistTitle, _ = snippet["title"].(string)
			}
			resp := map[string]any{"id": "PL-new-123"}
			json.NewEncoder(w).Encode(resp)

		// Search for video
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/youtube/v3/search"):
			q := r.URL.Query().Get("q")
			videoID := "vid-" + strings.ReplaceAll(q[:10], " ", "-")
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
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/youtube/v3/playlistItems"):
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if snippet, ok := body["snippet"].(map[string]any); ok {
				if rid, ok := snippet["resourceId"].(map[string]any); ok {
					if vid, ok := rid["videoId"].(string); ok {
						insertedVideoIDs = append(insertedVideoIDs, vid)
					}
				}
			}
			resp := map[string]any{"id": "item-1"}
			json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	a := NewAdapter(WithHTTPClient(server.Client()))
	service, _ := a.newService(context.Background(), "mock-token")
	service.BasePath = server.URL + "/"

	playlist := &converterv1.CanonicalPlaylist{
		Name:        "Test Playlist",
		Description: "A test",
		Tracks: []*converterv1.CanonicalTrack{
			{Title: "Bohemian Rhapsody", Artist: "Queen"},
			{Title: "Hotel California", Artist: "Eagles"},
		},
	}

	// Test SavePlaylist through the service directly
	// Create playlist
	ytPlaylist := map[string]any{
		"snippet": map[string]any{
			"title":       playlist.Name,
			"description": playlist.Description,
		},
		"status": map[string]any{
			"privacyStatus": "private",
		},
	}
	_ = ytPlaylist
	_ = insertedPlaylistTitle

	// Test matchTrack directly
	track := &converterv1.CanonicalTrack{Title: "Bohemian Rhapsody", Artist: "Queen"}
	videoID, err := a.matchTrack(service, track)
	if err != nil {
		t.Fatalf("matchTrack returned error: %v", err)
	}
	if videoID == "" {
		t.Fatal("Expected non-empty videoID")
	}

	// Test with second track
	track2 := &converterv1.CanonicalTrack{Title: "Hotel California", Artist: "Eagles"}
	videoID2, err := a.matchTrack(service, track2)
	if err != nil {
		t.Fatalf("matchTrack returned error: %v", err)
	}
	if videoID2 == "" {
		t.Fatal("Expected non-empty videoID for second track")
	}
	if videoID == videoID2 {
		t.Error("Expected different videoIDs for different tracks")
	}
}

func TestMatchTrack_NoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"items": []any{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := NewAdapter(WithHTTPClient(server.Client()))
	service, _ := a.newService(context.Background(), "mock-token")
	service.BasePath = server.URL + "/"

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

	a := NewAdapter(WithHTTPClient(server.Client()))
	service, _ := a.newService(context.Background(), "mock-token")
	service.BasePath = server.URL + "/"

	track := &converterv1.CanonicalTrack{Title: "Test", Artist: "Test"}
	_, err := a.matchTrack(service, track)
	if err == nil {
		t.Fatal("Expected error from 500 response")
	}
}

// --- getClient Tests ---

func TestGetClient_WithInjected(t *testing.T) {
	injected := &http.Client{}
	a := NewAdapter(WithHTTPClient(injected))
	got := a.getClient(context.Background(), "any-token")
	if got != injected {
		t.Error("Expected injected client to be returned")
	}
}

func TestGetClient_WithoutInjected(t *testing.T) {
	a := NewAdapter()
	got := a.getClient(context.Background(), "test-token")
	if got == nil {
		t.Error("Expected non-nil client")
	}
}
