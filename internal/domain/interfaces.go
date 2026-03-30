// Package domain defines the core abstractions for Portify's playlist conversion engine.
//
// If you are implementing a new streaming service adapter, start here.
// Your adapter must implement either PlaylistSource (to read playlists from a service)
// or PlaylistSink (to write playlists to a service) — or both.
package domain

import (
	"context"

	converterv1 "github.com/debalin/portify/gen/go/converter/v1"
)

// ProviderInfo contains basic metadata about a streaming service provider.
// Every adapter must return this from its Info() method.
type ProviderInfo struct {
	// ID is the unique, lowercase identifier for this provider (e.g. "spotify", "youtube").
	// This is used as the key in the provider registry and in RPC messages.
	ID string

	// Name is the human-readable display name (e.g. "Spotify", "YouTube Music").
	Name string

	// AuthURLHint is the full OAuth authorization URL that the frontend uses to
	// redirect users for authentication. It includes scopes, redirect URI, etc.
	AuthURLHint string
}

// PlaylistSource defines the interface for reading playlists from a streaming service.
//
// Implement this interface to add a new source platform (e.g., Spotify, Apple Music).
// The conversion engine calls these methods in order:
//  1. GetAuthURL() → user authorizes → frontend sends back the auth code
//  2. ExchangeAuthCode() → backend exchanges code for an access token
//  3. ListPlaylists() → user picks which playlist to convert
//  4. FetchPlaylist() → backend fetches the full playlist with all tracks
type PlaylistSource interface {
	// Info returns metadata about this provider (ID, display name, auth URL).
	// This is called during provider registration and when listing available providers.
	Info() ProviderInfo

	// GetAuthURL returns the full OAuth authorization URL for this provider.
	// The frontend opens this URL in the browser so the user can grant access.
	GetAuthURL() string

	// ExchangeAuthCode exchanges a one-time OAuth authorization code for an access token.
	// The authorization code comes from the OAuth callback after the user approves access.
	// Returns the access token string that should be used in subsequent API calls.
	ExchangeAuthCode(ctx context.Context, code string) (string, error)

	// ListPlaylists returns a summary of all playlists owned by the authenticated user.
	// The returned playlists contain ID, name, and description — but NOT full track lists.
	// Use FetchPlaylist() to get the complete track listing for a specific playlist.
	ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error)

	// FetchPlaylist retrieves a single playlist by ID, including ALL tracks with full metadata
	// (title, artist, album, duration, ISRC). This is the main method that powers conversion —
	// the tracks returned here are mapped to the destination service.
	// Large playlists should be fetched with pagination internally.
	FetchPlaylist(ctx context.Context, playlistID string, authToken string) (*converterv1.CanonicalPlaylist, error)
}

// PlaylistSink defines the interface for writing playlists to a streaming service.
//
// Implement this interface to add a new destination platform (e.g., YouTube Music, Tidal).
// The conversion engine calls these methods in order:
//  1. GetAuthURL() → user authorizes → frontend sends back the auth code
//  2. ExchangeAuthCode() → backend exchanges code for an access token
//  3. ListPlaylists() → user optionally picks an existing playlist to append to
//  4. SavePlaylist() → backend creates/updates the playlist with matched tracks
type PlaylistSink interface {
	// Info returns metadata about this provider (ID, display name, auth URL).
	Info() ProviderInfo

	// GetAuthURL returns the full OAuth authorization URL for this provider.
	GetAuthURL() string

	// ExchangeAuthCode exchanges a one-time OAuth authorization code for an access token.
	ExchangeAuthCode(ctx context.Context, code string) (string, error)

	// ListPlaylists returns a summary of all playlists owned by the authenticated user.
	// This is used so the user can optionally select an existing playlist to append to,
	// instead of always creating a new one.
	ListPlaylists(ctx context.Context, authToken string) ([]*converterv1.CanonicalPlaylist, error)

	// SavePlaylist creates a new playlist (or appends to an existing one) on this service.
	//
	// Parameters:
	//   - playlist: the canonical playlist with tracks to save
	//   - authToken: the user's access token for this service
	//   - destinationPlaylistID: if non-empty, tracks are appended to this existing playlist
	//     instead of creating a new one
	//   - onProgress: optional callback invoked after each track is processed, reporting
	//     cumulative counts of successfully converted and failed tracks
	//
	// Returns:
	//   - string: URL of the created/updated playlist on this service
	//   - []*CanonicalTrack: list of tracks that could not be matched or inserted
	//   - error: fatal error (e.g., auth failure, playlist creation failure)
	//
	// Per-track failures (no match found, insert error) are NOT fatal — the method
	// continues processing remaining tracks and reports failures via the return value.
	SavePlaylist(ctx context.Context, playlist *converterv1.CanonicalPlaylist, authToken string, destinationPlaylistID string, onProgress func(converted, failed int)) (string, []*converterv1.CanonicalTrack, error)
}
