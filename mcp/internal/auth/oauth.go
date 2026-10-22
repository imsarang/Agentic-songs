package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Login(ctx context.Context, path, id, secret, redirect string) error {
	u, err := url.Parse(redirect)
	if err != nil {
		return err
	}
	codeCh := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		codeCh <- r.URL.Query().Get("code")
		fmt.Fprint(w, "ok, close tab")
	})
	// Spotify redirects the host browser to 127.0.0.1; in Docker we must
	// listen on all interfaces so published port 4381 can reach us.
	listenAddr := u.Host
	if os.Getenv("SPOTIFY_IN_DOCKER") != "" {
		port := u.Port()
		if port == "" {
			port = "4381"
		}
		listenAddr = ":" + port
	}
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	q := url.Values{
		"client_id":     {id},
		"response_type": {"code"},
		"redirect_uri":  {redirect},
		"scope":         {"user-read-playback-state user-modify-playback-state user-read-currently-playing playlist-read-private playlist-modify-private playlist-modify-public"},
		"show_dialog":   {"true"}, // force consent so new scopes are not skipped
	}
	link := "https://accounts.spotify.com/authorize?" + q.Encode()
	log.Printf("login: %s", link)
	log.Printf("waiting for callback on %s (redirect_uri=%s)", listenAddr, redirect)
	_ = exec.Command("xdg-open", link).Start()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case code := <-codeCh:
		if code == "" {
			return fmt.Errorf("no code")
		}
		tok, err := tokenPost(ctx, id, secret, url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {code},
			"redirect_uri": {redirect},
		})
		if err != nil {
			return err
		}
		return Save(path, tok)
	case <-time.After(3 * time.Minute):
		return fmt.Errorf("login timeout")
	}
}

func tokenPost(ctx context.Context, id, secret string, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(id, secret)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return Token{}, fmt.Errorf("token %d: %s", res.StatusCode, b)
	}
	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return Token{}, err
	}
	return Token{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second),
	}, nil
}
