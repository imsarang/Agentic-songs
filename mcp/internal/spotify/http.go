package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"shanks/internal/auth"
)

type HTTPClient struct {
	http   *http.Client
	tokens *auth.Source
}

func NewHTTPClient(t *auth.Source) *HTTPClient {
	return &HTTPClient{http: &http.Client{Timeout: 20 * time.Second}, tokens: t}
}

func (c *HTTPClient) EnsureAuth(ctx context.Context) error { return c.tokens.EnsureReady(ctx) }

func (c *HTTPClient) Search(ctx context.Context, query string, limit int) ([]Hit, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 8 {
		limit = 8
	}
	q := url.Values{"q": {query}, "type": {"track"}, "limit": {strconv.Itoa(limit)}}
	var out struct {
		Tracks struct {
			Items []struct {
				Name, URI string
				Artists   []struct{ Name string }
			}
		}
	}
	if err := c.do(ctx, "GET", "/search?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	var hits []Hit
	for i, it := range out.Tracks.Items {
		a := ""
		if len(it.Artists) > 0 {
			a = it.Artists[0].Name
		}
		hits = append(hits, Hit{Index: i, Title: it.Name, Artist: a, URI: it.URI})
	}
	if len(hits) == 0 {
		return nil, fmt.Errorf("no tracks")
	}
	return hits, nil
}

func (c *HTTPClient) Play(ctx context.Context, uri string) error {
	return c.do(ctx, "PUT", "/me/player/play", map[string]any{"uris": []string{uri}}, nil)
}

func (c *HTTPClient) Control(ctx context.Context, action string) error {
	switch action {
	case "pause":
		return c.do(ctx, "PUT", "/me/player/pause", nil, nil)
	case "resume":
		return c.do(ctx, "PUT", "/me/player/play", nil, nil)
	case "next":
		return c.do(ctx, "POST", "/me/player/next", nil, nil)
	case "previous":
		return c.do(ctx, "POST", "/me/player/previous", nil, nil)
	default:
		return fmt.Errorf("bad action")
	}
}

func (c *HTTPClient) NowPlaying(ctx context.Context) (Bar, error) {
	var out struct {
		Item *struct {
			Name, URI string
			Artists   []struct{ Name string }
		}
	}
	if err := c.do(ctx, "GET", "/me/player/currently-playing", nil, &out); err != nil {
		return Bar{}, err
	}
	if out.Item == nil {
		return Bar{}, nil
	}
	a := ""
	if len(out.Item.Artists) > 0 {
		a = out.Item.Artists[0].Name
	}
	return Bar{Title: out.Item.Name, Artist: a, URI: out.Item.URI}, nil
}

func (c *HTTPClient) do(ctx context.Context, method, path string, in, out any) error {
	tok, err := c.tokens.AccessToken(ctx)
	if err != nil {
		return err
	}
	var body io.Reader
	if in != nil {
		b, _ := json.Marshal(in)
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.spotify.com/v1"+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode == 204 {
		return nil
	}
	if res.StatusCode == 404 {
		return fmt.Errorf("no active device; open Spotify once")
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("%s %s: %d %s", method, path, res.StatusCode, b)
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

var _ Client = (*HTTPClient)(nil)
