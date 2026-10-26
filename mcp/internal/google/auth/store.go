package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func Path(dataDir string) string {
	return filepath.Join(dataDir, "google-token.json")
}

func Load(path string) (Token, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Token{}, err
	}
	var t Token
	if err := json.Unmarshal(raw, &t); err != nil {
		return Token{}, err
	}
	return t, nil
}

func Save(path string, t Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}
