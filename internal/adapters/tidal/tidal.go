package tidal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/common"
	"github.com/debalin/portify/internal/domain"
	"golang.org/x/oauth2"
)

// Adapter implements domain.PlaylistSource and domain.PlaylistSink for Tidal
type Adapter struct {
	common.BaseAdapter
	BaseURL  string
	verifier string
	mu       sync.Mutex
}

// NewAdapter creates a new Tidal adapter instance
func NewAdapter(opts ...common.Option) *Adapter {
	a := &Adapter{
		BaseAdapter: common.BaseAdapter{
			OAuthCfg: common.OAuthConfig{
				ProviderID:   "tidal",
				ClientIDEnv:  "TIDAL_ID",
				ClientSecEnv: "TIDAL_SECRET",
				Scopes:       []string{"playlists.read", "playlists.write"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://login.tidal.com/authorize",
					TokenURL: "https://auth.tidal.com/v1/oauth2/token",
				},
			},
		},
		BaseURL: "https://openapi.tidal.com/v2",
	}
	a.ApplyOptions(opts)
	// If base URL was overridden by common options
	if a.BaseAdapter.BaseURL != "" {
		a.BaseURL = a.BaseAdapter.BaseURL
	}
	return a
}

// Info returns basic information about the Tidal provider
func (a *Adapter) Info() domain.ProviderInfo {
	return domain.ProviderInfo{
		ID:          "tidal",
		Name:        "Tidal",
		AuthURLHint: "", // Avoid calling GetAuthURL here to prevent verifier overwrite
	}
}

// GetAuthURL overrides the default to inject PKCE code challenge
func (a *Adapter) GetAuthURL() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.verifier = oauth2.GenerateVerifier()
	return a.GetOAuth2Config().AuthCodeURL(a.OAuthCfg.ProviderID, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(a.verifier))
}

// ExchangeAuthCode overrides the default to inject PKCE code verifier
func (a *Adapter) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	a.mu.Lock()
	verifier := a.verifier
	a.mu.Unlock()

	if verifier == "" {
		return "", fmt.Errorf("no PKCE verifier found for this session")
	}

	cfg := a.GetOAuth2Config()
	token, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return "", fmt.Errorf("oauth2 exchange failed: %w", err)
	}

	return token.AccessToken, nil
}

// helper to make JSON:API requests
func (a *Adapter) doRequest(ctx context.Context, authToken, method, path string, body io.Reader) ([]byte, error) {
	reqURL := a.BaseURL + path
	if strings.HasPrefix(path, "http") {
		reqURL = path
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Accept", "application/vnd.api+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}

	client := a.GetHTTPClient(ctx, authToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tidal api error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ListPlaylists fetches the user's existing Tidal playlists.
func (a *Adapter) ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error) {
	// Use the correct v2 endpoint discovered from Tidal API spec
	// /playlists?filter[owners.id]=me fetches playlists created by the user
	endpoint := "/playlists?filter[owners.id]=me"

	resp, err := a.doRequest(ctx, authToken, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tidal playlists: %w", err)
	}

	// Basic parsing of JSON:API
	var data struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("failed to parse tidal response: %w", err)
	}

	var canonicals []*converterv1.CanonicalPlaylist
	for _, item := range data.Data {
		canonicals = append(canonicals, &converterv1.CanonicalPlaylist{
			Id:          item.ID,
			Name:        item.Attributes.Name,
			Description: item.Attributes.Description,
		})
	}

	return canonicals, nil
}

// FetchPlaylist retrieves a single playlist by ID
func (a *Adapter) FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error) {
	// Request tracks and their associated artists
	resp, err := a.doRequest(ctx, authToken, http.MethodGet, "/playlists/"+playlistID+"?include=items,items.artists", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist metadata: %w", err)
	}

	var data struct {
		Data struct {
			Attributes struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"attributes"`
		} `json:"data"`
		Included []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				Title string `json:"title"` // For tracks
				Name  string `json:"name"`  // For artists
			} `json:"attributes"`
			Relationships struct {
				Artists struct {
					Data []struct {
						ID string `json:"id"`
					} `json:"data"`
				} `json:"artists"`
			} `json:"relationships"`
		} `json:"included"`
	}

	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, fmt.Errorf("failed to parse tidal response: %w", err)
	}

	canonical := &converterv1.CanonicalPlaylist{
		Name:        data.Data.Attributes.Name,
		Description: data.Data.Attributes.Description,
		Tracks:      make([]*converterv1.CanonicalTrack, 0),
	}

	// 1. Build an artist map
	artistsMap := make(map[string]string)
	for _, item := range data.Included {
		if item.Type == "artists" {
			artistsMap[item.ID] = item.Attributes.Name
		}
	}

	// 2. Map tracks
	for _, item := range data.Included {
		if item.Type == "tracks" {
			var artistNames []string
			for _, aData := range item.Relationships.Artists.Data {
				if name, ok := artistsMap[aData.ID]; ok {
					artistNames = append(artistNames, name)
				}
			}

			artistStr := "Unknown Artist"
			if len(artistNames) > 0 {
				artistStr = strings.Join(artistNames, ", ")
			}

			canonical.Tracks = append(canonical.Tracks, &converterv1.CanonicalTrack{
				Title:  item.Attributes.Title,
				Artist: artistStr,
			})
		}
	}

	return canonical, nil
}

func (a *Adapter) CreatePlaylist(ctx context.Context, name string, description string, authToken string) (string, error) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "playlists",
			"attributes": map[string]string{
				"name":        name,
				"description": description,
			},
		},
	}
	body, _ := json.Marshal(payload)
	resp, err := a.doRequest(ctx, authToken, http.MethodPost, "/playlists", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create tidal playlist: %w", err)
	}

	var data struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return "", fmt.Errorf("failed to parse created playlist response: %w", err)
	}
	return data.Data.ID, nil
}

func (a *Adapter) MatchTrack(ctx context.Context, track *converterv1.CanonicalTrack, authToken string) (string, error) {
	query := url.QueryEscape(fmt.Sprintf("%s %s", track.Title, track.Artist))
	endpoint := fmt.Sprintf("/searchResults/%s?include=tracks&countryCode=US", query)
	resp, err := a.doRequest(ctx, authToken, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}

	var data struct {
		Included []struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"included"`
	}

	if err := json.Unmarshal(resp, &data); err != nil {
		return "", fmt.Errorf("failed to parse search results: %w", err)
	}

	for _, item := range data.Included {
		if item.Type == "tracks" && item.ID != "" {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("track not found on Tidal")
}

func (a *Adapter) AddTrackToPlaylist(ctx context.Context, playlistID string, trackID string, authToken string) error {
	payload := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"type": "tracks",
				"id":   trackID,
			},
		},
	}
	body, _ := json.Marshal(payload)
	endpoint := fmt.Sprintf("/playlists/%s/relationships/items", playlistID)
	_, err := a.doRequest(ctx, authToken, http.MethodPost, endpoint, bytes.NewReader(body))
	return err
}

func (a *Adapter) GetPlaylistURL(playlistID string) string {
	return fmt.Sprintf("https://tidal.com/browse/playlist/%s", playlistID)
}
