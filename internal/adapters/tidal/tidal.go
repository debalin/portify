package tidal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/common"
	"github.com/debalin/portify/internal/domain"
	"golang.org/x/oauth2"
)

// Adapter implements domain.PlaylistSource and domain.PlaylistSink for Tidal
type Adapter struct {
	common.BaseAdapter
	BaseURL string
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
		AuthURLHint: a.GetAuthURL(),
	}
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
	// Attempting to fetch current user's playlists
	// Open API v2 usually uses /users/me/playlists or just /playlists with a filter
	// We'll try /playlists?filter[type]=USER
	resp, err := a.doRequest(ctx, authToken, http.MethodGet, "/playlists?include=items", nil)
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
	resp, err := a.doRequest(ctx, authToken, http.MethodGet, "/playlists/"+playlistID+"?include=items", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist metadata: %w", err)
	}

	var data struct {
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"attributes"`
		} `json:"data"`
		Included []struct {
			Type       string `json:"type"`
			Attributes struct {
				Title string `json:"title"`
				// Tidal JSON:API might have artists nested
			} `json:"attributes"`
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

	// We will likely need to fetch tracks properly if include=items isn't fully detailed
	return canonical, nil
}

func (a *Adapter) CreatePlaylist(ctx context.Context, name string, description string, authToken string) (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (a *Adapter) MatchTrack(ctx context.Context, track *converterv1.CanonicalTrack, authToken string) (string, error) {
	// Search API
	query := url.QueryEscape(fmt.Sprintf("%s %s", track.Title, track.Artist))
	resp, err := a.doRequest(ctx, authToken, http.MethodGet, "/search?query="+query+"&type=TRACKS", nil)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (a *Adapter) AddTrackToPlaylist(ctx context.Context, playlistID string, trackID string, authToken string) error {
	return fmt.Errorf("not implemented yet")
}

func (a *Adapter) GetPlaylistURL(playlistID string) string {
	return fmt.Sprintf("https://tidal.com/browse/playlist/%s", playlistID)
}
