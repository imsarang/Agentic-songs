package auth

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"
)

type Source struct {
	mu                               sync.Mutex
	path, clientID, secret, redirect string
	token                            Token
}

func NewSource(path, id, secret, redirect string) *Source {
	return &Source{path: path, clientID: id, secret: secret, redirect: redirect}
}

func (s *Source) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := Load(s.path)
	if err != nil {
		return err
	}
	s.token = t
	return nil
}

func (s *Source) EnsureReady(ctx context.Context) error {
	if err := s.Load(); err != nil {
		if !os.IsNotExist(err) {
			// ignore corrupt; login again
		}
		if err := Login(ctx, s.path, s.clientID, s.secret, s.redirect); err != nil {
			return err
		}
		return s.Load()
	}
	_, err := s.AccessToken(ctx)
	return err
}

func (s *Source) AccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token.AccessToken == "" {
		return "", fmt.Errorf("no token")
	}
	if time.Now().Before(s.token.Expiry.Add(-30 * time.Second)) {
		return s.token.AccessToken, nil
	}
	next, err := tokenPost(ctx, s.clientID, s.secret, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {s.token.RefreshToken},
	})
	if err != nil {
		return "", err
	}
	if next.RefreshToken == "" {
		next.RefreshToken = s.token.RefreshToken
	}
	s.token = next
	_ = Save(s.path, s.token)
	return s.token.AccessToken, nil
}
