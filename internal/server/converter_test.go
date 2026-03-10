package server

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/mock"
	"github.com/debalin/portify/internal/domain"
)

func TestConvertPlaylist_Success(t *testing.T) {
	// 1. Setup Registry with Mocks
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	registry.RegisterDestination(&mock.MockDestination{})

	// 2. Initialize Server
	srv := NewConverterServer(registry)

	// 3. Create Request
	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:       "spotify",
		DestinationProvider:  "youtube",
		SourcePlaylistId:     "mock-playlist-123",
		SourceAuthToken:      "source-token",
		DestinationAuthToken: "dest-token",
	})

	// 4. Call ConvertPlaylist
	res, err := srv.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !res.Msg.Success {
		t.Fatalf("Expected success to be true, got false. Message: %s", res.Msg.Message)
	}

	expectedURL := "https://youtube.com/playlist?list=mock"
	if res.Msg.DestinationPlaylistUrl != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, res.Msg.DestinationPlaylistUrl)
	}
}

func TestConvertPlaylist_SourceNotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	srv := NewConverterServer(registry)

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:      "invalid",
		DestinationProvider: "youtube",
	})

	_, err := srv.ConvertPlaylist(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing source provider, got none")
	}
}

func TestConvertPlaylist_DestinationNotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	srv := NewConverterServer(registry)

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:      "spotify",
		DestinationProvider: "invalid",
	})

	_, err := srv.ConvertPlaylist(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing destination provider, got none")
	}
}
