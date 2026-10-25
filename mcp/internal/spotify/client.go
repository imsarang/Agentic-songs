package spotify

import "context"

// Client is the Spotify Web API surface used by MCP tools.
type Client interface {
	EnsureAuth(ctx context.Context) error
	Search(ctx context.Context, query string, limit int) ([]Hit, error)
	Play(ctx context.Context, uri string) error
	PlayPlaylist(ctx context.Context, playlistURI string) error
	Control(ctx context.Context, action string) error
	NowPlaying(ctx context.Context) (Bar, error)
	ListPlaylists(ctx context.Context) ([]Playlist, error)
	CreatePlaylist(ctx context.Context, name, description string) (Playlist, error)
	AddToPlaylist(ctx context.Context, playlistID, trackURI string) error
	UpdatePlaylistMetadata(ctx context.Context, playlistID, name, description string) error
	DeletePlaylist(ctx context.Context, playlistID string) error
	RemoveFromPlaylist(ctx context.Context, playlistID, trackURI string) error
}
