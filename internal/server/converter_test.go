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

// ======== ListProviders ========

func TestListProviders_Empty(t *testing.T) {
	registry := domain.NewProviderRegistry()
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ListProviders(context.Background(), connect.NewRequest(&converterv1.ListProvidersRequest{}))
	if err != nil {
		t.Fatalf("ListProviders returned error: %v", err)
	}
	if len(res.Msg.Sources) != 0 {
		t.Errorf("Expected 0 sources, got %d", len(res.Msg.Sources))
	}
	if len(res.Msg.Destinations) != 0 {
		t.Errorf("Expected 0 destinations, got %d", len(res.Msg.Destinations))
	}
}

func TestListProviders_WithProviders(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	registry.RegisterDestination(&mock.MockDestination{})

	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ListProviders(context.Background(), connect.NewRequest(&converterv1.ListProvidersRequest{}))
	if err != nil {
		t.Fatalf("ListProviders returned error: %v", err)
	}
	if len(res.Msg.Sources) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(res.Msg.Sources))
	}
	if res.Msg.Sources[0].Id != "spotify" {
		t.Errorf("Expected source ID 'spotify', got '%s'", res.Msg.Sources[0].Id)
	}
	if len(res.Msg.Destinations) != 1 {
		t.Fatalf("Expected 1 destination, got %d", len(res.Msg.Destinations))
	}
	if res.Msg.Destinations[0].Id != "youtube" {
		t.Errorf("Expected destination ID 'youtube', got '%s'", res.Msg.Destinations[0].Id)
	}
}

// ======== GetAuthURL ========

func TestGetAuthURL_Source(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.GetAuthURL(context.Background(), connect.NewRequest(&converterv1.GetAuthURLRequest{
		ProviderId: "spotify",
	}))
	if err != nil {
		t.Fatalf("GetAuthURL returned error: %v", err)
	}
	if res.Msg.AuthUrl == "" {
		t.Error("Expected non-empty auth URL")
	}
}

func TestGetAuthURL_Destination(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterDestination(&mock.MockDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.GetAuthURL(context.Background(), connect.NewRequest(&converterv1.GetAuthURLRequest{
		ProviderId: "youtube",
	}))
	if err != nil {
		t.Fatalf("GetAuthURL returned error: %v", err)
	}
	if res.Msg.AuthUrl == "" {
		t.Error("Expected non-empty auth URL")
	}
}

func TestGetAuthURL_NotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	_, err := client.GetAuthURL(context.Background(), connect.NewRequest(&converterv1.GetAuthURLRequest{
		ProviderId: "nonexistent",
	}))
	if err == nil {
		t.Fatal("Expected error for nonexistent provider")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// ======== ExchangeAuthCode ========

func TestExchangeAuthCode_Source(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ExchangeAuthCode(context.Background(), connect.NewRequest(&converterv1.ExchangeAuthCodeRequest{
		ProviderId: "spotify",
		Code:       "test-code",
	}))
	if err != nil {
		t.Fatalf("ExchangeAuthCode returned error: %v", err)
	}
	if !res.Msg.Success {
		t.Error("Expected success=true")
	}
	if res.Msg.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
}

func TestExchangeAuthCode_Destination(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterDestination(&mock.MockDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ExchangeAuthCode(context.Background(), connect.NewRequest(&converterv1.ExchangeAuthCodeRequest{
		ProviderId: "youtube",
		Code:       "test-code",
	}))
	if err != nil {
		t.Fatalf("ExchangeAuthCode returned error: %v", err)
	}
	if !res.Msg.Success {
		t.Error("Expected success=true")
	}
	if res.Msg.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
}

func TestExchangeAuthCode_NotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	_, err := client.ExchangeAuthCode(context.Background(), connect.NewRequest(&converterv1.ExchangeAuthCodeRequest{
		ProviderId: "nonexistent",
		Code:       "test-code",
	}))
	if err == nil {
		t.Fatal("Expected error for nonexistent provider")
	}
}

// ======== ListUserPlaylists ========

func TestListUserPlaylists_Source(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ListUserPlaylists(context.Background(), connect.NewRequest(&converterv1.ListUserPlaylistsRequest{
		ProviderId:  "spotify",
		AccessToken: "mock-token",
	}))
	if err != nil {
		t.Fatalf("ListUserPlaylists returned error: %v", err)
	}
	if len(res.Msg.Playlists) != 2 {
		t.Errorf("Expected 2 playlists from mock source, got %d", len(res.Msg.Playlists))
	}
}

func TestListUserPlaylists_Destination(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterDestination(&mock.MockDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ListUserPlaylists(context.Background(), connect.NewRequest(&converterv1.ListUserPlaylistsRequest{
		ProviderId:  "youtube",
		AccessToken: "mock-token",
	}))
	if err != nil {
		t.Fatalf("ListUserPlaylists returned error: %v", err)
	}
	if len(res.Msg.Playlists) != 1 {
		t.Errorf("Expected 1 playlist from mock destination, got %d", len(res.Msg.Playlists))
	}
}

func TestListUserPlaylists_NotFound(t *testing.T) {
	registry := domain.NewProviderRegistry()
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	_, err := client.ListUserPlaylists(context.Background(), connect.NewRequest(&converterv1.ListUserPlaylistsRequest{
		ProviderId:  "nonexistent",
		AccessToken: "mock-token",
	}))
	if err == nil {
		t.Fatal("Expected error for nonexistent provider")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// ======== ConvertPlaylist ========

func TestConvertPlaylist_Success(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSource{})
	registry.RegisterDestination(&mock.MockDestination{})

	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:       "spotify",
		DestinationProvider:  "youtube",
		SourcePlaylistId:     "mock-playlist-123",
		SourceAuthToken:      "source-token",
		DestinationAuthToken: "dest-token",
	})

	stream, err := client.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	var finalRes *converterv1.ConvertPlaylistResponse
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

func TestConvertPlaylist_WithTracks(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSourceWithTracks{})
	registry.RegisterDestination(&mock.MockDestination{})

	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:       "spotify",
		DestinationProvider:  "youtube",
		SourcePlaylistId:     "playlist-with-tracks",
		SourceAuthToken:      "source-token",
		DestinationAuthToken: "dest-token",
	})

	stream, err := client.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	statusCounts := map[converterv1.ConvertPlaylistResponse_Status]int{}
	var finalRes *converterv1.ConvertPlaylistResponse

	for stream.Receive() {
		msg := stream.Msg()
		statusCounts[msg.Status]++
		if msg.Status == converterv1.ConvertPlaylistResponse_STATUS_DONE {
			finalRes = msg
		}
	}

	if err := stream.Err(); err != nil {
		t.Fatalf("Expected no stream error, got: %v", err)
	}

	// Should have FETCHING, CONVERTING, and DONE statuses
	if statusCounts[converterv1.ConvertPlaylistResponse_STATUS_FETCHING] != 1 {
		t.Error("Expected exactly 1 FETCHING status")
	}
	if statusCounts[converterv1.ConvertPlaylistResponse_STATUS_DONE] != 1 {
		t.Error("Expected exactly 1 DONE status")
	}

	if finalRes == nil {
		t.Fatal("Expected DONE message")
	}
	if finalRes.TracksTotal != 3 {
		t.Errorf("Expected 3 total tracks, got %d", finalRes.TracksTotal)
	}
}

func TestConvertPlaylist_WithAppendToExisting(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSourceWithTracks{})
	registry.RegisterDestination(&mock.MockDestination{})

	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	req := connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:        "spotify",
		DestinationProvider:   "youtube",
		SourcePlaylistId:      "playlist-with-tracks",
		DestinationPlaylistId: "existing-dest-playlist-123",
		SourceAuthToken:       "source-token",
		DestinationAuthToken:  "dest-token",
	})

	stream, err := client.ConvertPlaylist(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	var finalRes *converterv1.ConvertPlaylistResponse
	for stream.Receive() {
		msg := stream.Msg()
		if msg.Status == converterv1.ConvertPlaylistResponse_STATUS_DONE {
			finalRes = msg
		}
	}

	if err := stream.Err(); err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	if finalRes == nil {
		t.Fatal("Expected DONE message")
	}
}

func TestNewConverterServer(t *testing.T) {
	registry := domain.NewProviderRegistry()
	srv := NewConverterServer(registry)
	if srv == nil {
		t.Fatal("Expected non-nil server")
	}
	if srv.registry != registry {
		t.Error("Expected server to hold the provided registry")
	}
}

// ======== Error Path Tests ========

func TestExchangeAuthCode_SourceError(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockFailingSource{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ExchangeAuthCode(context.Background(), connect.NewRequest(&converterv1.ExchangeAuthCodeRequest{
		ProviderId: "failing-source",
		Code:       "test-code",
	}))
	if err != nil {
		t.Fatalf("ExchangeAuthCode returned RPC error: %v", err)
	}
	if res.Msg.Success {
		t.Error("Expected success=false for failing source")
	}
	if res.Msg.ErrorMessage == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestExchangeAuthCode_DestinationError(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterDestination(&mock.MockFailingDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	res, err := client.ExchangeAuthCode(context.Background(), connect.NewRequest(&converterv1.ExchangeAuthCodeRequest{
		ProviderId: "failing-dest",
		Code:       "test-code",
	}))
	if err != nil {
		t.Fatalf("ExchangeAuthCode returned RPC error: %v", err)
	}
	if res.Msg.Success {
		t.Error("Expected success=false for failing destination")
	}
	if res.Msg.ErrorMessage == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestListUserPlaylists_SourceError(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockFailingSource{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	_, err := client.ListUserPlaylists(context.Background(), connect.NewRequest(&converterv1.ListUserPlaylistsRequest{
		ProviderId:  "failing-source",
		AccessToken: "token",
	}))
	if err == nil {
		t.Fatal("Expected error from failing source ListPlaylists")
	}
}

func TestListUserPlaylists_DestinationError(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterDestination(&mock.MockFailingDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	_, err := client.ListUserPlaylists(context.Background(), connect.NewRequest(&converterv1.ListUserPlaylistsRequest{
		ProviderId:  "failing-dest",
		AccessToken: "token",
	}))
	if err == nil {
		t.Fatal("Expected error from failing destination ListPlaylists")
	}
}

func TestConvertPlaylist_FetchPlaylistError(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockFailingSource{})
	registry.RegisterDestination(&mock.MockDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	stream, err := client.ConvertPlaylist(context.Background(), connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:       "failing-source",
		DestinationProvider:  "youtube",
		SourcePlaylistId:     "any",
		SourceAuthToken:      "token",
		DestinationAuthToken: "token",
	}))
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	var gotError bool
	for stream.Receive() {
		if stream.Msg().Status == converterv1.ConvertPlaylistResponse_STATUS_ERROR {
			gotError = true
			if !strings.Contains(stream.Msg().Message, "fetch") {
				t.Errorf("Expected fetch error message, got: %s", stream.Msg().Message)
			}
		}
	}
	if !gotError {
		t.Error("Expected an ERROR status in the stream for fetch failure")
	}
}

func TestConvertPlaylist_SavePlaylistError(t *testing.T) {
	registry := domain.NewProviderRegistry()
	registry.RegisterSource(&mock.MockSourceWithTracks{})
	registry.RegisterDestination(&mock.MockFailingDestination{})
	ts, client := setupTestServer(t, registry)
	defer ts.Close()

	stream, err := client.ConvertPlaylist(context.Background(), connect.NewRequest(&converterv1.ConvertPlaylistRequest{
		SourceProvider:       "spotify",
		DestinationProvider:  "failing-dest",
		SourcePlaylistId:     "any",
		SourceAuthToken:      "token",
		DestinationAuthToken: "token",
	}))
	if err != nil {
		t.Fatalf("Expected no error initializing stream, got: %v", err)
	}

	var gotError bool
	for stream.Receive() {
		if stream.Msg().Status == converterv1.ConvertPlaylistResponse_STATUS_ERROR {
			gotError = true
			if !strings.Contains(stream.Msg().Message, "save") || !strings.Contains(stream.Msg().Message, "destination") {
				t.Errorf("Expected save/destination error message, got: %s", stream.Msg().Message)
			}
		}
	}
	if !gotError {
		t.Error("Expected an ERROR status in the stream for save failure")
	}
}
