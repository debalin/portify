package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"connectrpc.com/connect"
	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
	"github.com/debalin/portify/internal/adapters/common"
	"github.com/debalin/portify/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
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

// ConvertPlaylist orchestrates the full conversion flow and streams progress back to the client.
//
// The flow is:
//  1. Fetch playlist from source (FetchPlaylist)
//  2. Create or select destination playlist (CreatePlaylist or use existing ID)
//  3. For each track: match on destination (MatchTrack) + insert (AddTrackToPlaylist)
//  4. Stream progress after each track
func (s *ConverterServer) ConvertPlaylist(
	ctx context.Context,
	req *connect.Request[converterv1.ConvertPlaylistRequest],
	stream *connect.ServerStream[converterv1.ConvertPlaylistResponse],
) error {
	log.Printf("Request received: ConvertPlaylist from %s to %s", req.Msg.SourceProvider, req.Msg.DestinationProvider)

	ctx, span := common.Tracer.Start(ctx, "ConvertPlaylist")
	defer span.End()

	span.SetAttributes(
		attribute.String("source_provider", req.Msg.SourceProvider),
		attribute.String("destination_provider", req.Msg.DestinationProvider),
	)

	conversionStatus := "success"
	defer func() {
		common.ConversionsTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("source", req.Msg.SourceProvider),
			attribute.String("destination", req.Msg.DestinationProvider),
			attribute.String("status", conversionStatus),
		))
		if conversionStatus == "failed" {
			span.SetStatus(codes.Error, "conversion failed")
		} else {
			span.SetStatus(codes.Ok, "conversion completed successfully")
		}
	}()

	source, ok := s.registry.GetSource(req.Msg.SourceProvider)
	if !ok {
		conversionStatus = "failed"
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("source provider %s not found", req.Msg.SourceProvider))
	}

	dest, ok := s.registry.GetDestination(req.Msg.DestinationProvider)
	if !ok {
		conversionStatus = "failed"
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("destination provider %s not found", req.Msg.DestinationProvider))
	}

	// Step 1: Fetch playlist from source
	stream.Send(&converterv1.ConvertPlaylistResponse{
		Status:  converterv1.ConvertPlaylistResponse_STATUS_FETCHING,
		Message: "Fetching playlist data from source...",
	})

	fetchCtx, fetchSpan := common.Tracer.Start(ctx, fmt.Sprintf("FetchPlaylist: %s", req.Msg.SourceProvider))
	canonicalPlaylist, err := source.FetchPlaylist(fetchCtx, req.Msg.SourcePlaylistId, req.Msg.SourceAuthToken)
	if err != nil {
		fetchSpan.RecordError(err)
		fetchSpan.SetStatus(codes.Error, err.Error())
		fetchSpan.End()

		conversionStatus = "failed"
		stream.Send(&converterv1.ConvertPlaylistResponse{
			Status:  converterv1.ConvertPlaylistResponse_STATUS_ERROR,
			Message: fmt.Sprintf("Failed to fetch source playlist: %v", err),
		})
		return nil
	}
	fetchSpan.SetAttributes(
		attribute.Int("tracks_count", len(canonicalPlaylist.Tracks)),
	)
	fetchSpan.End()

	totalTracks := int32(len(canonicalPlaylist.Tracks))

	// Step 2: Create or select destination playlist
	playlistID := req.Msg.DestinationPlaylistId
	var existingTracks []*converterv1.CanonicalTrack
	if playlistID == "" {
		createCtx, createSpan := common.Tracer.Start(ctx, fmt.Sprintf("CreatePlaylist: %s", req.Msg.DestinationProvider))
		playlistID, err = dest.CreatePlaylist(createCtx, canonicalPlaylist.Name, canonicalPlaylist.Description, req.Msg.DestinationAuthToken)
		if err != nil {
			createSpan.RecordError(err)
			createSpan.SetStatus(codes.Error, err.Error())
			createSpan.End()

			conversionStatus = "failed"
			stream.Send(&converterv1.ConvertPlaylistResponse{
				Status:  converterv1.ConvertPlaylistResponse_STATUS_ERROR,
				Message: fmt.Sprintf("Failed to create destination playlist: %v", err),
			})
			return nil
		}
		createSpan.SetAttributes(
			attribute.String("playlist_id", playlistID),
		)
		createSpan.End()
	} else {
		// If appending to an existing playlist, fetch existing tracks to deduplicate and skip matching
		if destSource, ok := s.registry.GetSource(req.Msg.DestinationProvider); ok {
			stream.Send(&converterv1.ConvertPlaylistResponse{
				Status:      converterv1.ConvertPlaylistResponse_STATUS_FETCHING,
				Message:     "Checking existing tracks in destination playlist to avoid duplicates...",
				TracksTotal: totalTracks,
			})

			fetchDestCtx, fetchDestSpan := common.Tracer.Start(ctx, fmt.Sprintf("FetchExistingTracks: %s", req.Msg.DestinationProvider))
			existingPlaylist, err := destSource.FetchPlaylist(fetchDestCtx, playlistID, req.Msg.DestinationAuthToken)
			if err != nil {
				log.Printf("Warning: could not fetch existing tracks for playlist %s: %v", playlistID, err)
				fetchDestSpan.RecordError(err)
			} else if existingPlaylist != nil {
				existingTracks = existingPlaylist.Tracks
				fetchDestSpan.SetAttributes(attribute.Int("existing_tracks_count", len(existingTracks)))
			}
			fetchDestSpan.End()
		}
	}

	stream.Send(&converterv1.ConvertPlaylistResponse{
		Status:      converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
		Message:     fmt.Sprintf("Starting conversion for '%s'...", canonicalPlaylist.Name),
		TracksTotal: totalTracks,
	})

	// Step 3: Match and insert each track
	converted := int32(0)
	failed := int32(0)
	skipped := int32(0)
	var failedTracks []*converterv1.CanonicalTrack

	loopCtx, loopSpan := common.Tracer.Start(ctx, "Match & Insert Tracks")
	defer loopSpan.End()

	for _, track := range canonicalPlaylist.Tracks {
		trackCtx, trackSpan := common.Tracer.Start(loopCtx, fmt.Sprintf("ProcessTrack: %s - %s", track.Artist, track.Title))
		trackSpan.SetAttributes(
			attribute.String("track.title", track.Title),
			attribute.String("track.artist", track.Artist),
		)

		// Check if track already exists in destination playlist to save API quota and avoid duplicates
		if domain.TrackExistsInList(track, existingTracks) {
			skipped++
			trackSpan.SetAttributes(attribute.String("track.status", "skipped_duplicate"))
			trackSpan.End()

			common.TracksProcessedTotal.Add(ctx, 1, metric.WithAttributes(
				attribute.String("provider", req.Msg.DestinationProvider),
				attribute.String("status", "skipped"),
			))

			progressMsg := fmt.Sprintf("Converting tracks... (%d/%d)", converted+failed+skipped, totalTracks)
			if skipped > 0 {
				progressMsg = fmt.Sprintf("Converting tracks... (%d/%d, %d skipped)", converted+failed+skipped, totalTracks, skipped)
			}

			stream.Send(&converterv1.ConvertPlaylistResponse{
				Status:          converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
				Message:         progressMsg,
				TracksTotal:     totalTracks,
				TracksConverted: converted,
				TracksFailed:    failed,
				TracksSkipped:   skipped,
			})
			continue
		}

		retryHook := func(event common.RetryEvent) {
			stream.Send(&converterv1.ConvertPlaylistResponse{
				Status: converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
				Message: fmt.Sprintf("[%s] Rate limited (HTTP %d). Retrying in %v (attempt %d)...",
					strings.ToUpper(event.ProviderID), event.StatusCode, event.Delay.Round(time.Second), event.Attempt),
				TracksTotal:     totalTracks,
				TracksConverted: converted,
				TracksFailed:    failed,
				TracksSkipped:   skipped,
			})
		}
		trackCtx = common.WithRetryHook(trackCtx, retryHook)

		// Match
		trackID, err := dest.MatchTrack(trackCtx, track, req.Msg.DestinationAuthToken)
		if err != nil || trackID == "" {
			failed++
			failedTracks = append(failedTracks, track)

			trackSpan.SetAttributes(attribute.String("track.status", "match_failed"))
			if err != nil {
				trackSpan.RecordError(err)
			}
			trackSpan.End()

			common.TracksProcessedTotal.Add(ctx, 1, metric.WithAttributes(
				attribute.String("provider", req.Msg.DestinationProvider),
				attribute.String("status", "match_failed"),
			))

			progressMsg := fmt.Sprintf("Converting tracks... (%d/%d)", converted+failed+skipped, totalTracks)
			if skipped > 0 {
				progressMsg = fmt.Sprintf("Converting tracks... (%d/%d, %d skipped)", converted+failed+skipped, totalTracks, skipped)
			}

			stream.Send(&converterv1.ConvertPlaylistResponse{
				Status:          converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
				Message:         progressMsg,
				TracksTotal:     totalTracks,
				TracksConverted: converted,
				TracksFailed:    failed,
				TracksSkipped:   skipped,
			})
			continue
		}

		trackSpan.SetAttributes(attribute.String("destination_track_id", trackID))

		// Insert
		err = dest.AddTrackToPlaylist(trackCtx, playlistID, trackID, req.Msg.DestinationAuthToken)
		trackStatus := "success"
		if err != nil {
			failed++
			failedTracks = append(failedTracks, track)
			trackStatus = "insert_failed"
			trackSpan.RecordError(err)
			trackSpan.SetStatus(codes.Error, err.Error())
		} else {
			converted++
		}
		trackSpan.SetAttributes(attribute.String("track.status", trackStatus))
		trackSpan.End()

		common.TracksProcessedTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("provider", req.Msg.DestinationProvider),
			attribute.String("status", trackStatus),
		))

		progressMsg := fmt.Sprintf("Converting tracks... (%d/%d)", converted+failed+skipped, totalTracks)
		if skipped > 0 {
			progressMsg = fmt.Sprintf("Converting tracks... (%d/%d, %d skipped)", converted+failed+skipped, totalTracks, skipped)
		}

		stream.Send(&converterv1.ConvertPlaylistResponse{
			Status:          converterv1.ConvertPlaylistResponse_STATUS_CONVERTING,
			Message:         progressMsg,
			TracksTotal:     totalTracks,
			TracksConverted: converted,
			TracksFailed:    failed,
			TracksSkipped:   skipped,
		})
	}

	// Step 4: Done
	doneMessage := fmt.Sprintf("Successfully converted '%s'.", canonicalPlaylist.Name)
	if skipped > 0 {
		doneMessage = fmt.Sprintf("Successfully converted '%s'. %d added, %d already in playlist.", canonicalPlaylist.Name, converted, skipped)
	}

	stream.Send(&converterv1.ConvertPlaylistResponse{
		Status:                 converterv1.ConvertPlaylistResponse_STATUS_DONE,
		Message:                doneMessage,
		DestinationPlaylistUrl: dest.GetPlaylistURL(playlistID),
		TracksTotal:            totalTracks,
		TracksConverted:        converted,
		TracksFailed:           failed,
		TracksSkipped:          skipped,
		FailedTracks:           failedTracks,
	})

	return nil
}
