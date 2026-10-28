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

const (
	authURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenURL = "https://oauth2.googleapis.com/token"
	// users.messages.list, users.drafts.create, users.drafts.send, users.messages.trash
	// people.searchContacts
	// drive.about.get
	scopes = "https://www.googleapis.com/auth/gmail.modify https://www.googleapis.com/auth/contacts.readonly https://www.googleapis.com/auth/drive"
)

func Login(ctx context.Context, path, id, secret, redirect string) error {
	u, err := url.Parse(redirect)
	if err != nil {
		return err
	}
	codeChannel := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		codeChannel <- r.URL.Query().Get("code")
		fmt.Fprint(w, "ok, close tab")
	})

	listenAddr := u.Host
	if os.Getenv("GOOGLE_IN_DOCKER") != "" {
		port := u.Port()
		if port == "" {
			port = "4382"
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
		"scope":         {scopes},
		"access_type":   {"offline"},
		"prompt":        {"select_account consent"},
	}
	link := authURL + "?" + q.Encode()
	log.Printf("login: %s", link)
	log.Printf("waiting for callback on %s (redirect_uri=%s)", listenAddr, redirect)
	openBrowser(ctx, link)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case code := <-codeChannel:
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

func openBrowser(ctx context.Context, link string) {
	if os.Getenv("GOOGLE_IN_DOCKER") == "" {
		if err := exec.Command("xdg-open", link).Start(); err != nil {
			log.Printf("login: xdg-open: %v", err)
		}
		return
	}
	endpoint := os.Getenv("GOOGLE_OPEN_URL")
	if endpoint == "" {
		endpoint = "http://host.docker.internal:4390/open"
	}
	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, strings.NewReader(link))
	if err != nil {
		log.Printf("login: browser opener: %v", err)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("login: browser opener unreachable: %v", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		log.Printf("login: browser opener status %d", res.StatusCode)
		return
	}
	log.Printf("login: asked the host to open the browser")
}

func tokenPost(ctx context.Context, id, secret string, form url.Values) (Token, error) {
	form.Set("client_id", id)
	form.Set("client_secret", secret)
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
