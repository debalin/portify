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
			Id:          source.ID,
			Name:        source.Name,
			AuthUrlHint: source.AuthURLHint,
		})
	}

	var destinations []*converterv1.ProviderInfo
	for _, dest := range s.registry.ListDestinations() {
		destinations = append(destinations, &converterv1.ProviderInfo{
			Id:          dest.ID,
			Name:        dest.Name,
			AuthUrlHint: dest.AuthURLHint,
		})
	}

	res := connect.NewResponse(&converterv1.ListProvidersResponse{
		Sources:      sources,
		Destinations: destinations,
	})

	return res, nil
}

// GetAuthURL triggers the generation of an OAuth login URL.
func (s *ConverterServer) GetAuthURL(
	ctx context.Context,
	req *connect.Request[converterv1.GetAuthURLRequest],
) (*connect.Response[converterv1.GetAuthURLResponse], error) {
	log.Printf("Request received: GetAuthURL for %s", req.Msg.ProviderId)

	// Try checking sources first
	if source, ok := s.registry.GetSource(req.Msg.ProviderId); ok {
		return connect.NewResponse(&converterv1.GetAuthURLResponse{
			AuthUrl: source.GetAuthURL(),
		}), nil
	}

	// Try checking destinations
	if dest, ok := s.registry.GetDestination(req.Msg.ProviderId); ok {
		return connect.NewResponse(&converterv1.GetAuthURLResponse{
			AuthUrl: dest.GetAuthURL(),
		}), nil
	}

	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("provider %s not found", req.Msg.ProviderId))
}

// ExchangeAuthCode trades the callback code for a token.
func (s *ConverterServer) ExchangeAuthCode(
	ctx context.Context,
	req *connect.Request[converterv1.ExchangeAuthCodeRequest],
) (*connect.Response[converterv1.ExchangeAuthCodeResponse], error) {
	log.Printf("Request received: ExchangeAuthCode for %s", req.Msg.ProviderId)

	if source, ok := s.registry.GetSource(req.Msg.ProviderId); ok {
		token, err := source.ExchangeAuthCode(ctx, req.Msg.Code)
		if err != nil {
			return connect.NewResponse(&converterv1.ExchangeAuthCodeResponse{
				Success:      false,
				ErrorMessage: err.Error(),
			}), nil
		}
		return connect.NewResponse(&converterv1.ExchangeAuthCodeResponse{
			Success:     true,
			AccessToken: token,
		}), nil
	}

	if dest, ok := s.registry.GetDestination(req.Msg.ProviderId); ok {
		token, err := dest.ExchangeAuthCode(ctx, req.Msg.Code)
		if err != nil {
			return connect.NewResponse(&converterv1.ExchangeAuthCodeResponse{
				Success:      false,
				ErrorMessage: err.Error(),
			}), nil
		}
		return connect.NewResponse(&converterv1.ExchangeAuthCodeResponse{
			Success:     true,
			AccessToken: token,
		}), nil
	}

	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("provider %s not found", req.Msg.ProviderId))
}

// ListUserPlaylists returns the playlists available to a user on a given platform.
func (s *ConverterServer) ListUserPlaylists(
	ctx context.Context,
	req *connect.Request[converterv1.ListUserPlaylistsRequest],
) (*connect.Response[converterv1.ListUserPlaylistsResponse], error) {
	log.Printf("Request received: ListUserPlaylists for %s", req.Msg.ProviderId)

	source, ok1 := s.registry.GetSource(req.Msg.ProviderId)
	if ok1 {
		playlists, err := source.ListPlaylists(ctx, req.Msg.AccessToken)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		var protoPlaylists []*converterv1.CanonicalPlaylist
		protoPlaylists = append(protoPlaylists, playlists...)
		return connect.NewResponse(&converterv1.ListUserPlaylistsResponse{
			Playlists: protoPlaylists,
		}), nil
	}

	dest, ok2 := s.registry.GetDestination(req.Msg.ProviderId)
	if ok2 {
		playlists, err := dest.ListPlaylists(ctx, req.Msg.AccessToken)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		var protoPlaylists []*converterv1.CanonicalPlaylist
		protoPlaylists = append(protoPlaylists, playlists...)
		return connect.NewResponse(&converterv1.ListUserPlaylistsResponse{
			Playlists: protoPlaylists,
		}), nil
	}

	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("provider %s not found", req.Msg.ProviderId))
}

// ConvertPlaylist triggers the conversion process and streams progress back.
func (s *ConverterServer) ConvertPlaylist(
	ctx context.Context,
	req *connect.Request[converterv1.ConvertPlaylistRequest],
	stream *connect.ServerStream[converterv1.ConvertPlaylistResponse],
) error {
	log.Printf("Request received: ConvertPlaylist from %s to %s", req.Msg.SourceProvider, req.Msg.DestinationProvider)

	source, ok := s.registry.GetSource(req.Msg.SourceProvider)
	if !ok {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("source provider %s not found", req.Msg.SourceProvider))
	}

	dest, ok := s.registry.GetDestination(req.Msg.DestinationProvider)
	if !ok {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("destination provider %s not found", req.Msg.DestinationProvider))
	}

	// Send initial "FETCHING" status
	stream.Send(&converterv1.ConvertPlaylistResponse{
		Status:  converterv1.ConvertPlaylistResponse_STATUS_FETCHING,
		Message: "Fetching playlist data from source...",
	})

	// 1. Fetch playlist from source
	canonicalPlaylist, err := source.FetchPlaylist(ctx, req.Msg.SourcePlaylistId, req.Msg.SourceAuthToken)
	if err != nil {
		stream.Send(&converterv1.ConvertPlaylistResponse{
			Status:  converterv1.ConvertPlaylistResponse_STATUS_ERROR,
			Message: fmt.Sprintf("Failed to fetch source playlist: %v", err),
		})
		return nil
	}

	totalTracks := int32(len(canonicalPlaylist.Tracks))

	// Send "CONVERTING" status start
	stream.Send(&converterv1.ConvertPlaylistResponse{
		Status:      converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
		Message:     fmt.Sprintf("Starting conversion for '%s'...", canonicalPlaylist.Name),
		TracksTotal: totalTracks,
	})

	var convertedFinal, failedFinal int

	// 2. Save playlist to destination with progress callback
	playlistURL, failedTracks, err := dest.SavePlaylist(
		ctx,
		canonicalPlaylist,
		req.Msg.DestinationAuthToken,
		req.Msg.DestinationPlaylistId,
		func(converted, failed int) {
			convertedFinal = converted
			failedFinal = failed
			stream.Send(&converterv1.ConvertPlaylistResponse{
				Status:          converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
				Message:         fmt.Sprintf("Converting tracks... (%d/%d)", converted+failed, totalTracks),
				TracksTotal:     totalTracks,
				TracksConverted: int32(converted),
				TracksFailed:    int32(failed),
			})
		},
	)

	if err != nil {
		stream.Send(&converterv1.ConvertPlaylistResponse{
			Status:  converterv1.ConvertPlaylistResponse_STATUS_ERROR,
			Message: fmt.Sprintf("Failed to save playlist to destination: %v", err),
		})
		return nil
	}

	// Send "DONE" status
	stream.Send(&converterv1.ConvertPlaylistResponse{
		Status:                 converterv1.ConvertPlaylistResponse_STATUS_DONE,
		Message:                fmt.Sprintf("Successfully converted '%s'.", canonicalPlaylist.Name),
		DestinationPlaylistUrl: playlistURL,
		TracksTotal:            totalTracks,
		TracksConverted:        int32(convertedFinal),
		TracksFailed:           int32(failedFinal),
		FailedTracks:           failedTracks,
	})

	return nil
}
