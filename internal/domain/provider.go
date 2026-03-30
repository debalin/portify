package domain

// ProviderRegistry manages the available source and sink adapters.
// Sources and sinks are registered at startup by main.go and looked up by the RPC server.
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
