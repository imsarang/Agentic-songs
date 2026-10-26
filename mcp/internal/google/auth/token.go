package auth

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"
)

type Source struct {
	mu                               sync.Mutex
	path, clientId, secret, redirect string
	token                            Token
}

func NewSource(path, id, secret, redirect string) *Source {
	return &Source{
		path:     path,
		clientId: id,
		secret:   secret,
		redirect: redirect,
	}
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
		if err := Login(ctx, s.path, s.clientId, s.secret, s.redirect); err != nil {
			return err
		}
		return s.Load()
	}
	_, err := s.AccessToken(ctx)
	return err
}

func (s *Source) ReLogin(ctx context.Context) error {
	s.mu.Lock()
	s.token = Token{}
	path := s.path
	id, secret, redirect := s.clientId, s.secret, s.redirect
	s.mu.Unlock()
	if err := Login(ctx, path, id, secret, redirect); err != nil {
		_ = s.Load()
		return err
	}
	return s.Load()
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
	next, err := tokenPost(ctx, s.clientId, s.secret, url.Values{
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
