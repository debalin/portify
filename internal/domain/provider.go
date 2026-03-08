package domain

import (
	"context"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
)

// ProviderInfo contains basic information about a provider
type ProviderInfo struct {
	ID   string
	Name string
}

// PlaylistSource defines the interface for fetching playlists from a source platform (e.g., Spotify)
type PlaylistSource interface {
	Info() ProviderInfo
	FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error)
}

// PlaylistSink defines the interface for creating/saving playlists to a destination platform (e.g., YouTube Music)
type PlaylistSink interface {
	Info() ProviderInfo
	SavePlaylist(ctx context.Context, playlist *converterv1.CanonicalPlaylist, authToken string) (string, error) // Returns destination URL
}

// TrackMatcher defines the interface for matching a canonical track on a specific platform
type TrackMatcher interface {
	Match(ctx context.Context, track *converterv1.CanonicalTrack) (string, error) // Returns platform-specific track ID
}

// ProviderRegistry manages the available sources and sinks
type ProviderRegistry struct {
	sources      map[string]PlaylistSource
	destinations map[string]PlaylistSink
}

// NewProviderRegistry creates a new registry instance
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		sources:      make(map[string]PlaylistSource),
		destinations: make(map[string]PlaylistSink),
	}
}

// RegisterSource adds a new source provider to the registry
func (r *ProviderRegistry) RegisterSource(source PlaylistSource) {
	r.sources[source.Info().ID] = source
}

// RegisterDestination adds a new destination provider to the registry
func (r *ProviderRegistry) RegisterDestination(sink PlaylistSink) {
	r.destinations[sink.Info().ID] = sink
}

// GetSource retrieves a source provider by its ID
func (r *ProviderRegistry) GetSource(id string) (PlaylistSource, bool) {
	source, ok := r.sources[id]
	return source, ok
}

// GetDestination retrieves a destination provider by its ID
func (r *ProviderRegistry) GetDestination(id string) (PlaylistSink, bool) {
	sink, ok := r.destinations[id]
	return sink, ok
}

// ListSources returns information about all registered sources
func (r *ProviderRegistry) ListSources() []ProviderInfo {
	var infos []ProviderInfo
	for _, s := range r.sources {
		infos = append(infos, s.Info())
	}
	return infos
}

// ListDestinations returns information about all registered destinations
func (r *ProviderRegistry) ListDestinations() []ProviderInfo {
	var infos []ProviderInfo
	for _, d := range r.destinations {
		infos = append(infos, d.Info())
	}
	return infos
}
