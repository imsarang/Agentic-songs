package cache

import (
	"context"
	"testing"

	"shanks/internal/spotify"
)

type fakeClient struct {
	hits []spotify.Hit
}

func (f *fakeClient) EnsureAuth(context.Context) error { return nil }
func (f *fakeClient) Search(context.Context, string, int) ([]spotify.Hit, error) {
	return f.hits, nil
}
func (f *fakeClient) Play(context.Context, string) error                      { return nil }
func (f *fakeClient) PlayPlaylist(context.Context, string) error              { return nil }
func (f *fakeClient) Control(context.Context, string) error                   { return nil }
func (f *fakeClient) NowPlaying(context.Context) (spotify.NowPlaying, error) {
	return spotify.NowPlaying{}, nil
}
func (f *fakeClient) ListPlaylists(context.Context) ([]spotify.Playlist, error) {
	return nil, nil
}
func (f *fakeClient) CreatePlaylist(context.Context, string, string) (spotify.Playlist, error) {
	return spotify.Playlist{}, nil
}
func (f *fakeClient) AddToPlaylist(context.Context, string, string) error { return nil }
func (f *fakeClient) UpdatePlaylistMetadata(context.Context, string, string, string) error {
	return nil
}
func (f *fakeClient) DeletePlaylist(context.Context, string) error            { return nil }
func (f *fakeClient) RemoveFromPlaylist(context.Context, string, string) error { return nil }
func (f *fakeClient) AudioFeatures(context.Context, string) (spotify.AudioFeatures, error) {
	return spotify.AudioFeatures{}, nil
}

func TestPickPrefersMatchingTitle(t *testing.T) {
	f := &fakeClient{hits: []spotify.Hit{
		{Index: 0, Title: "Tu Tu Hai Wahi - Remix", Artist: "V", URI: "wrong"},
		{Index: 1, Title: "Bin Tere Sanam", Artist: "Udit", URI: "right"},
		{Index: 2, Title: "Zara Zara Touch Me", Artist: "M", URI: "other"},
	}}
	var s Search
	hit, err := s.Pick(context.Background(), f, "Bin Tere Sanam", 0, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if hit.URI != "right" {
		t.Fatalf("got %+v", hit)
	}
}
