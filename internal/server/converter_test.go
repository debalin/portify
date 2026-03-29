package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/gen/go/converter/v1/converterv1connect"
	"github.com/debalin/portify/internal/adapters/mock"
	"github.com/debalin/portify/internal/domain"
)

func setupTestServer(t *testing.T, registry *domain.ProviderRegistry) (*httptest.Server, converterv1connect.ConverterServiceClient) {
	srv := NewConverterServer(registry)
	mux := http.NewServeMux()
	path, handler := converterv1connect.NewConverterServiceHandler(srv)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)

	client := converterv1connect.NewConverterServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	return testServer, client
}

func TestConvertPlaylist_Success(t *testing.T) {
	// 1. Setup Registry with Mocks
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	registry.RegisterDestination(&mock.MockDestination{})

	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	// 2. Create Request
	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:       "spotify",
		DestinationProvider:  "youtube",
		SourcePlaylistId:     "mock-playlist-123",
		SourceAuthToken:      "source-token",
		DestinationAuthToken: "dest-token",
	})

	// 3. Call ConvertPlaylist
	stream, err := client.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	var finalRes *converterv1.ConvertPlaylistResponse
	// 4. Consume stream
	for stream.Receive() {
		msg := stream.Msg()
		t.Logf("Received message: Status=%v, Message=%v", msg.Status, msg.Message)
		if msg.Status == converterv1.ConvertPlaylistResponse_STATUS_DONE {
			finalRes = msg
		} else if msg.Status == converterv1.ConvertPlaylistResponse_STATUS_ERROR {
			t.Fatalf("Received error status in stream: %s", msg.Message)
		}
	}

	if err := stream.Err(); err != nil {
		t.Fatalf("Expected no stream error, got: %v", err)
	}

	if finalRes == nil {
		t.Fatalf("Expected DONE status, but stream ended before receiving it")
	}

	expectedURL := "https://youtube.com/playlist?list=mock"
	if finalRes.DestinationPlaylistUrl != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, finalRes.DestinationPlaylistUrl)
	}
}

func TestConvertPlaylist_SourceNotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:      "invalid",
		DestinationProvider: "youtube",
	})

	stream, err := client.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	for stream.Receive() {
		if stream.Msg().Status == converterv1.ConvertPlaylistResponse_STATUS_ERROR {
			// Some tests might send status error explicitly instead of go return err
		}
	}

	err = stream.Err()
	if err == nil {
		t.Fatal("Expected error for missing source provider, got none")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestConvertPlaylist_DestinationNotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:      "spotify",
		DestinationProvider: "invalid",
	})

	stream, err := client.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	for stream.Receive() {
	}

	err = stream.Err()
	if err == nil {
		t.Fatal("Expected error for missing destination provider, got none")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected not found error, got: %v", err)
	}
}
