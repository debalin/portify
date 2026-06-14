package tidal

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

// --- Helper mux for Tidal API simulation ---

func tidalMux(
	createdPlaylistID string,
	searchResults map[string]string, // query substring -> trackId
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")

		// ListPlaylists
		if r.Method == http.MethodGet && r.URL.Path == "/playlists" {
			resp := map[string]any{
				"data": []map[string]any{
					{
						"id": "111",
						"attributes": map[string]any{
							"name":        "Test Playlist",
							"description": "Desc",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// FetchPlaylist
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/playlists/") && !strings.Contains(r.URL.Path, "relationships") {
			resp := map[string]any{
				"data": map[string]any{
					"attributes": map[string]any{
						"name":        "Fetched Playlist",
						"description": "Fetched Desc",
					},
				},
				"included": []map[string]any{
					{
						"id":   "artist1",
						"type": "artists",
						"attributes": map[string]any{
							"name": "The Beatles",
						},
					},
					{
						"id":   "track1",
						"type": "tracks",
						"attributes": map[string]any{
							"title": "Hey Jude",
						},
						"relationships": map[string]any{
							"artists": map[string]any{
								"data": []map[string]any{
									{"id": "artist1"},
								},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// CreatePlaylist
		if r.Method == http.MethodPost && r.URL.Path == "/playlists" {
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"id": createdPlaylistID},
			})
			return
		}

		// Tracks ISRC endpoint simulation
		if r.Method == http.MethodGet && r.URL.Path == "/tracks" {
			isrc := r.URL.Query().Get("filter[isrc]")
			if isrc == "USUM71703861" {
				resp := map[string]any{
					"data": []map[string]any{
						{
							"id":   "isrc_track_123",
							"type": "tracks",
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
		}

		// MatchTrack (Search)
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/searchResults/") {
			query := strings.TrimPrefix(r.URL.Path, "/searchResults/")
			// Remove query params
			if idx := strings.Index(query, "?"); idx != -1 {
				query = query[:idx]
			}
			query = strings.ReplaceAll(query, "%20", " ")
			query = strings.ReplaceAll(query, "+", " ")

			for substr, trackID := range searchResults {
				if strings.Contains(strings.ToLower(query), strings.ToLower(substr)) {
					json.NewEncoder(w).Encode(map[string]any{
						"included": []map[string]any{
							{"type": "tracks", "id": trackID},
						},
					})
					return
				}
			}
			json.NewEncoder(w).Encode(map[string]any{"included": []any{}})
			return
		}

		// AddTrackToPlaylist
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/relationships/items") {
			w.WriteHeader(http.StatusCreated)
			return
		}

		http.NotFound(w, r)
	})
}

func newTestAdapter(serverURL string) *Adapter {
	return NewAdapter(
		common.WithHTTPClient(http.DefaultClient),
		common.WithBaseURL(serverURL),
	)
}

// --- Tests ---

func TestInfo(t *testing.T) {
	a := NewAdapter()
	info := a.Info()
	if info.ID != "tidal" {
		t.Errorf("Expected ID 'tidal', got '%s'", info.ID)
	}
	if info.Name != "Tidal" {
		t.Errorf("Expected Name 'Tidal', got '%s'", info.Name)
	}
}

func TestListPlaylists(t *testing.T) {
	server := httptest.NewServer(tidalMux("new_id", nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlists, err := a.ListPlaylists(context.Background(), "token")
	if err != nil {
		t.Fatalf("ListPlaylists failed: %v", err)
	}

	if len(playlists) != 1 {
		t.Fatalf("Expected 1 playlist, got %d", len(playlists))
	}
	if playlists[0].Id != "111" {
		t.Errorf("Expected ID 111, got %s", playlists[0].Id)
	}
	if playlists[0].Name != "Test Playlist" {
		t.Errorf("Expected Name 'Test Playlist', got %s", playlists[0].Name)
	}
}

func TestFetchPlaylist(t *testing.T) {
	server := httptest.NewServer(tidalMux("new_id", nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	playlist, err := a.FetchPlaylist(context.Background(), "111", "token")
	if err != nil {
		t.Fatalf("FetchPlaylist failed: %v", err)
	}

	if playlist.Name != "Fetched Playlist" {
		t.Errorf("Expected 'Fetched Playlist', got %s", playlist.Name)
	}
	if len(playlist.Tracks) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(playlist.Tracks))
	}

	track := playlist.Tracks[0]
	if track.Title != "Hey Jude" {
		t.Errorf("Expected Title 'Hey Jude', got %s", track.Title)
	}
	if track.Artist != "The Beatles" {
		t.Errorf("Expected Artist 'The Beatles', got %s", track.Artist)
	}
}

func TestCreatePlaylist(t *testing.T) {
	server := httptest.NewServer(tidalMux("new_playlist_id", nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	id, err := a.CreatePlaylist(context.Background(), "My New Playlist", "Desc", "token")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}
	if id != "new_playlist_id" {
		t.Errorf("Expected 'new_playlist_id', got %s", id)
	}
}

func TestMatchTrack(t *testing.T) {
	searchMock := map[string]string{
		"hey jude": "track999",
	}
	server := httptest.NewServer(tidalMux("new_id", searchMock))
	defer server.Close()

	a := newTestAdapter(server.URL)

	track := &converterv1.CanonicalTrack{Title: "Hey Jude", Artist: "The Beatles"}
	id, err := a.MatchTrack(context.Background(), track, "token")
	if err != nil {
		t.Fatalf("MatchTrack failed: %v", err)
	}
	if id != "track999" {
		t.Errorf("Expected 'track999', got %s", id)
	}

	// Test fallback/no match
	track2 := &converterv1.CanonicalTrack{Title: "Nonexistent", Artist: "Nobody"}
	id2, err := a.MatchTrack(context.Background(), track2, "token")
	if err == nil {
		t.Fatalf("Expected error for no match, got nil")
	}
	if id2 != "" {
		t.Errorf("Expected empty string for no match, got %s", id2)
	}
}

func TestAddTrackToPlaylist(t *testing.T) {
	server := httptest.NewServer(tidalMux("new_id", nil))
	defer server.Close()

	a := newTestAdapter(server.URL)
	err := a.AddTrackToPlaylist(context.Background(), "playlist123", "track999", "token")
	if err != nil {
		t.Fatalf("AddTrackToPlaylist failed: %v", err)
	}
}

// --- Error Tests for Coverage ---

func errorMux(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	})
}

func TestErrors_APIError(t *testing.T) {
	server := httptest.NewServer(errorMux(http.StatusBadRequest, `{"error":"bad request"}`))
	defer server.Close()
	a := newTestAdapter(server.URL)
	ctx := context.Background()

	_, err := a.ListPlaylists(ctx, "token")
	if err == nil {
		t.Error("ListPlaylists expected API error, got nil")
	}

	_, err = a.FetchPlaylist(ctx, "111", "token")
	if err == nil {
		t.Error("FetchPlaylist expected API error, got nil")
	}

	_, err = a.CreatePlaylist(ctx, "Name", "Desc", "token")
	if err == nil {
		t.Error("CreatePlaylist expected API error, got nil")
	}

	_, err = a.MatchTrack(ctx, &converterv1.CanonicalTrack{Title: "Title", Artist: "Artist"}, "token")
	if err == nil {
		t.Error("MatchTrack expected API error, got nil")
	}

	err = a.AddTrackToPlaylist(ctx, "pid", "tid", "token")
	if err == nil {
		t.Error("AddTrackToPlaylist expected API error, got nil")
	}
}

func TestErrors_JSONParse(t *testing.T) {
	server := httptest.NewServer(errorMux(http.StatusOK, `invalid json`))
	defer server.Close()
	a := newTestAdapter(server.URL)
	ctx := context.Background()

	_, err := a.ListPlaylists(ctx, "token")
	if err == nil {
		t.Error("ListPlaylists expected JSON parse error, got nil")
	}

	_, err = a.FetchPlaylist(ctx, "111", "token")
	if err == nil {
		t.Error("FetchPlaylist expected JSON parse error, got nil")
	}

	_, err = a.CreatePlaylist(ctx, "Name", "Desc", "token")
	if err == nil {
		t.Error("CreatePlaylist expected JSON parse error, got nil")
	}

	_, err = a.MatchTrack(ctx, &converterv1.CanonicalTrack{Title: "Title", Artist: "Artist"}, "token")
	if err == nil {
		t.Error("MatchTrack expected JSON parse error, got nil")
	}
}

func TestAuthFlow(t *testing.T) {
	a := NewAdapter()
	url := a.GetAuthURL()
	if !strings.Contains(url, "login.tidal.com") {
		t.Errorf("Expected tidal login URL, got %s", url)
	}

	// ExchangeAuthCode failure (invalid token / mock endpoint missing)
	server := httptest.NewServer(errorMux(http.StatusBadRequest, `{"error":"invalid_grant"}`))
	defer server.Close()
	a = newTestAdapter(server.URL)
	a.OAuthCfg.Endpoint.TokenURL = server.URL // Override for test

	_, err := a.ExchangeAuthCode(context.Background(), "dummy_code")
	if err == nil {
		t.Error("Expected ExchangeAuthCode error, got nil")
	}
}

func TestMatchTrack_ISRC_And_Fuzzy(t *testing.T) {
	server := httptest.NewServer(tidalMux("new_id", map[string]string{
		"hey jude": "track999",
	}))
	defer server.Close()
	a := newTestAdapter(server.URL)

	// Test exact ISRC match
	track := &converterv1.CanonicalTrack{
		Title:  "Hey Jude",
		Artist: "The Beatles",
		Isrc:   "USUM71703861",
	}
	id, err := a.MatchTrack(context.Background(), track, "token")
	if err != nil {
		t.Fatalf("ISRC MatchTrack failed: %v", err)
	}
	if id != "isrc_track_123" {
		t.Errorf("Expected 'isrc_track_123', got %s", id)
	}

	// Test fuzzy match (text fallback)
	trackNoIsrc := &converterv1.CanonicalTrack{
		Title:  "Hey Jude",
		Artist: "The Beatles",
	}
	id2, err := a.MatchTrack(context.Background(), trackNoIsrc, "token")
	if err != nil {
		t.Fatalf("Fuzzy MatchTrack failed: %v", err)
	}
	if id2 != "track999" {
		t.Errorf("Expected 'track999', got %s", id2)
	}
}
