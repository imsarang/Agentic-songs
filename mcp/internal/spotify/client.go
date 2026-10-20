package spotify

import "context"

// Client is the Spotify Web API surface used by the actor. No browser.
type Client interface {
	EnsureAuth(ctx context.Context) error
	Search(ctx context.Context, query string, limit int) ([]Hit, error)
	Play(ctx context.Context, uri string) error
	Control(ctx context.Context, action string) error
	NowPlaying(ctx context.Context) (Bar, error)
}
