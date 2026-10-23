package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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

// ensureDeviceID finds a Connect device (opens the web player if needed).
// Inactive devices are fine — play/control pass device_id to wake them.
func (c *HTTPClient) ensureDeviceID(ctx context.Context) (string, error) {
	if id, err := c.pickDeviceID(ctx); err != nil {
		return "", err
	} else if id != "" {
		return id, nil
	}
	log.Printf("no device; waiting for host Spotify (or open web player)")
	if os.Getenv("SPOTIFY_IN_DOCKER") == "" {
		if err := exec.Command("xdg-open", "https://open.spotify.com").Start(); err != nil {
			return "", fmt.Errorf("open web player: %w", err)
		}
	}
	// In Docker, make up starts desktop Spotify; poll Connect devices.
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
		}
		id, err := c.pickDeviceID(ctx)
		if err != nil {
			return "", err
		}
		if id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("no device; open open.spotify.com, log in, leave the tab open")
}

func (c *HTTPClient) pickDeviceID(ctx context.Context) (string, error) {
	var out struct {
		Devices []struct {
			ID       string `json:"id"`
			IsActive bool   `json:"is_active"`
		} `json:"devices"`
	}
	if err := c.do(ctx, "GET", "/me/player/devices", nil, &out); err != nil {
		return "", err
	}
	for _, d := range out.Devices {
		if d.IsActive && d.ID != "" {
			return d.ID, nil
		}
	}
	if len(out.Devices) > 0 && out.Devices[0].ID != "" {
		return out.Devices[0].ID, nil
	}
	return "", nil
}

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
	id, err := c.ensureDeviceID(ctx)
	if err != nil {
		return err
	}
	path := "/me/player/play?device_id=" + url.QueryEscape(id)
	return c.do(ctx, "PUT", path, map[string]any{"uris": []string{uri}}, nil)
}

// PlayPlaylist starts a playlist on the active device (Spotify queues the rest).
func (c *HTTPClient) PlayPlaylist(ctx context.Context, playlistURI string) error {
	if playlistURI == "" {
		return fmt.Errorf("playlist uri required")
	}
	id, err := c.ensureDeviceID(ctx)
	if err != nil {
		return err
	}
	path := "/me/player/play?device_id=" + url.QueryEscape(id)
	return c.do(ctx, "PUT", path, map[string]any{"context_uri": playlistURI}, nil)
}

func (c *HTTPClient) Control(ctx context.Context, action string) error {
	id, err := c.ensureDeviceID(ctx)
	if err != nil {
		return err
	}
	q := "?device_id=" + url.QueryEscape(id)
	switch action {
	case "pause":
		return c.do(ctx, "PUT", "/me/player/pause"+q, nil, nil)
	case "resume":
		return c.do(ctx, "PUT", "/me/player/play"+q, nil, nil)
	case "next":
		return c.do(ctx, "POST", "/me/player/next"+q, nil, nil)
	case "previous":
		return c.do(ctx, "POST", "/me/player/previous"+q, nil, nil)
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

func (c *HTTPClient) ListPlaylists(ctx context.Context) ([]Playlist, error) {
	var out struct {
		Items []struct {
			ID, Name, Description, URI string
		} `json:"items"`
	}
	if err := c.do(ctx, "GET", "/me/playlists?limit=50", nil, &out); err != nil {
		return nil, err
	}
	list := make([]Playlist, 0, len(out.Items))
	for _, it := range out.Items {
		list = append(list, Playlist{
			ID: it.ID, Name: it.Name, Description: it.Description, URI: it.URI,
		})
	}
	return list, nil
}

func (c *HTTPClient) CreatePlaylist(ctx context.Context, name, description string) (Playlist, error) {
	body := map[string]any{
		"name":        name,
		"description": description,
		"public":      false,
	}
	var out struct {
		ID, Name, Description, URI string
	}
	if err := c.do(ctx, "POST", "/me/playlists", body, &out); err != nil {
		return Playlist{}, err
	}
	return Playlist{ID: out.ID, Name: out.Name, Description: out.Description, URI: out.URI}, nil
}

func (c *HTTPClient) AddToPlaylist(ctx context.Context, playlistID, trackURI string) error {
	if playlistID == "" || trackURI == "" {
		return fmt.Errorf("playlist id and track uri required")
	}
	body := map[string]any{"uris": []string{trackURI}}
	// /tracks is rejected with 403 for Development Mode apps; /items is the current path.
	return c.do(ctx, "POST", "/playlists/"+playlistID+"/items", body, nil)
}

func (c *HTTPClient) UpdatePlaylistMetadata(ctx context.Context, playlistID, name, description string) error {
	if playlistID == "" {
		return fmt.Errorf("playlist id required")
	}
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if description != "" {
		body["description"] = description
	}
	if len(body) == 0 {
		return fmt.Errorf("name or description required")
	}
	return c.do(ctx, "PUT", "/playlists/"+playlistID, body, nil)
}

func (c *HTTPClient) DeletePlaylist(ctx context.Context, playlistID string) error {
	if playlistID == "" {
		return fmt.Errorf("playlist id required")
	}
	uri := "spotify:playlist:" + playlistID
	return c.do(ctx, "DELETE", "/me/library?uris="+url.QueryEscape(uri), nil, nil)
}

func (c *HTTPClient) RemoveFromPlaylist(ctx context.Context, playlistID, trackURI string) error {
	if playlistID == "" || trackURI == "" {
		return fmt.Errorf("playlist id and track uri required")
	}
	body := map[string]any{
		"items": []map[string]string{{"uri": trackURI}},
	}
	return c.do(ctx, "DELETE", "/playlists/"+playlistID+"/items", body, nil)
}

func (c *HTTPClient) AudioFeatures(ctx context.Context, trackID string) (TrackFeatures, error) {
	if trackID == "" {
		return TrackFeatures{}, fmt.Errorf("track id required")
	}
	var out TrackFeatures
	if err := c.do(ctx, "GET", "/audio-features/"+url.PathEscape(trackID), nil, &out); err != nil {
		return TrackFeatures{}, err
	}
	return out, nil
}

func (c *HTTPClient) do(ctx context.Context, method, path string, in, out any) error {
	for attempt := 0; attempt < 2; attempt++ {
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
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode == 204 {
			return nil
		}
		if res.StatusCode == 403 && attempt == 0 && bytes.Contains(bytes.ToLower(b), []byte("insufficient")) {
			log.Printf("insufficient scope; re-login")
			if err := c.tokens.Relogin(ctx); err != nil {
				return err
			}
			continue
		}
		if res.StatusCode >= 300 {
			return fmt.Errorf("%s %s: %d %s", method, path, res.StatusCode, b)
		}
		if out == nil || len(b) == 0 {
			return nil
		}
		return json.Unmarshal(b, out)
	}
	return fmt.Errorf("%s %s: retry failed", method, path)
}

var _ Client = (*HTTPClient)(nil)
