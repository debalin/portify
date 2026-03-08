package server

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/domain"
)

// ConverterServer implements the ConverterService API.
type ConverterServer struct {
	registry *domain.ProviderRegistry
}

// NewConverterServer creates a new instance of the server.
func NewConverterServer(registry *domain.ProviderRegistry) *ConverterServer {
	return &ConverterServer{
		registry: registry,
	}
}

// ListProviders returns a list of supported source and destination providers.
func (s *ConverterServer) ListProviders(
	ctx context.Context,
	req *connect.Request[converterv1.ListProvidersRequest],
) (*connect.Response[converterv1.ListProvidersResponse], error) {
	log.Println("Request received: ListProviders")

	var sources []*converterv1.ProviderInfo
	for _, source := range s.registry.ListSources() {
		sources = append(sources, &converterv1.ProviderInfo{
			Id:   source.ID,
			Name: source.Name,
		})
	}

	var destinations []*converterv1.ProviderInfo
	for _, dest := range s.registry.ListDestinations() {
		destinations = append(destinations, &converterv1.ProviderInfo{
			Id:   dest.ID,
			Name: dest.Name,
		})
	}

	res := connect.NewResponse(&converterv1.ListProvidersResponse{
		Sources:      sources,
		Destinations: destinations,
	})

	return res, nil
}

// ConvertPlaylist triggers the conversion process.
func (s *ConverterServer) ConvertPlaylist(
	ctx context.Context,
	req *connect.Request[converterv1.ConvertPlaylistRequest],
) (*connect.Response[converterv1.ConvertPlaylistResponse], error) {
	log.Printf("Request received: ConvertPlaylist from %s to %s", req.Msg.SourceProvider, req.Msg.DestinationProvider)

	// Mock response for now
	return connect.NewResponse(&converterv1.ConvertPlaylistResponse{
		Success:         true,
		Message:         fmt.Sprintf("Mock conversion from %s to %s completed", req.Msg.SourceProvider, req.Msg.DestinationProvider),
		TracksTotal:     10,
		TracksConverted: 10,
		TracksFailed:    0,
	}), nil
}
