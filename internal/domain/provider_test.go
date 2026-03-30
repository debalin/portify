package domain

import (
	"context"
	"testing"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
)

// --- Minimal stub implementations for testing ---

type stubSource struct {
	id string
}

func (s *stubSource) Info() ProviderInfo {
	return ProviderInfo{ID: s.id, Name: s.id + " name", AuthURLHint: "https://auth." + s.id}
}
func (s *stubSource) GetAuthURL() string { return "https://auth." + s.id }
func (s *stubSource) ExchangeAuthCode(_ context.Context, code string) (string, error) {
	return "token-" + code, nil
}
func (s *stubSource) ListPlaylists(_ context.Context, _ string) ([]*converterv1.CanonicalPlaylist, error) {
	return []*converterv1.CanonicalPlaylist{{Id: "pl-1", Name: "Playlist 1"}}, nil
}
func (s *stubSource) FetchPlaylist(_ context.Context, id string, _ string) (*converterv1.CanonicalPlaylist, error) {
	return &converterv1.CanonicalPlaylist{Id: id, Name: "Fetched " + id}, nil
}

type stubSink struct {
	id string
}

func (d *stubSink) Info() ProviderInfo {
	return ProviderInfo{ID: d.id, Name: d.id + " name", AuthURLHint: "https://auth." + d.id}
}
func (d *stubSink) GetAuthURL() string { return "https://auth." + d.id }
func (d *stubSink) ExchangeAuthCode(_ context.Context, code string) (string, error) {
	return "token-" + code, nil
}
func (d *stubSink) ListPlaylists(_ context.Context, _ string) ([]*converterv1.CanonicalPlaylist, error) {
	return []*converterv1.CanonicalPlaylist{{Id: "dest-1", Name: "Dest Playlist"}}, nil
}
func (d *stubSink) SavePlaylist(_ context.Context, _ *converterv1.CanonicalPlaylist, _ string, _ string, _ func(int, int)) (string, []*converterv1.CanonicalTrack, error) {
	return "https://dest.example.com/playlist", nil, nil
}

// --- Tests ---

func TestNewProviderRegistry(t *testing.T) {
	r := NewProviderRegistry()
	if r == nil {
		t.Fatal("Expected non-nil registry")
	}
	if len(r.ListSources()) != 0 {
		t.Error("Expected empty sources")
	}
	if len(r.ListDestinations()) != 0 {
		t.Error("Expected empty destinations")
	}
}

func TestRegisterAndGetSource(t *testing.T) {
	r := NewProviderRegistry()
	src := &stubSource{id: "spotify"}
	r.RegisterSource(src)

	got, ok := r.GetSource("spotify")
	if !ok {
		t.Fatal("Expected to find source 'spotify'")
	}
	if got.Info().ID != "spotify" {
		t.Errorf("Expected ID 'spotify', got %s", got.Info().ID)
	}

	_, ok = r.GetSource("nonexistent")
	if ok {
		t.Error("Expected false for nonexistent source")
	}
}

func TestRegisterAndGetDestination(t *testing.T) {
	r := NewProviderRegistry()
	sink := &stubSink{id: "youtube"}
	r.RegisterDestination(sink)

	got, ok := r.GetDestination("youtube")
	if !ok {
		t.Fatal("Expected to find destination 'youtube'")
	}
	if got.Info().ID != "youtube" {
		t.Errorf("Expected ID 'youtube', got %s", got.Info().ID)
	}

	_, ok = r.GetDestination("nonexistent")
	if ok {
		t.Error("Expected false for nonexistent destination")
	}
}

func TestListSources(t *testing.T) {
	r := NewProviderRegistry()
	r.RegisterSource(&stubSource{id: "spotify"})
	r.RegisterSource(&stubSource{id: "tidal"})

	sources := r.ListSources()
	if len(sources) != 2 {
		t.Fatalf("Expected 2 sources, got %d", len(sources))
	}

	ids := map[string]bool{}
	for _, s := range sources {
		ids[s.ID] = true
	}
	if !ids["spotify"] || !ids["tidal"] {
		t.Errorf("Expected spotify and tidal in sources, got %v", ids)
	}
}

func TestListDestinations(t *testing.T) {
	r := NewProviderRegistry()
	r.RegisterDestination(&stubSink{id: "youtube"})
	r.RegisterDestination(&stubSink{id: "apple"})

	dests := r.ListDestinations()
	if len(dests) != 2 {
		t.Fatalf("Expected 2 destinations, got %d", len(dests))
	}

	ids := map[string]bool{}
	for _, d := range dests {
		ids[d.ID] = true
	}
	if !ids["youtube"] || !ids["apple"] {
		t.Errorf("Expected youtube and apple in destinations, got %v", ids)
	}
}

func TestOverwriteSource(t *testing.T) {
	r := NewProviderRegistry()
	r.RegisterSource(&stubSource{id: "spotify"})
	r.RegisterSource(&stubSource{id: "spotify"}) // same ID

	sources := r.ListSources()
	if len(sources) != 1 {
		t.Errorf("Expected 1 source after re-register, got %d", len(sources))
	}
}
