package config

import (
	"bufio"
	"os"
	"strings"
)

const (
	defaultDataDir = "data"
	envClientID    = "GOOGLE_CLIENT_ID"
	envSecret      = "GOOGLE_CLIENT_SECRET"
	envRedirect    = "GOOGLE_REDIRECT_URI"
	envDataDir     = "GOOGLE_DATA_DIR"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	DataDir      string
}

func Load() (Config, error) {
	file, err := loadEnvFile()
	if err != nil {
		return Config{}, err
	}
	dataDir := pick(file, envDataDir)
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	redirect := pick(file, envRedirect)
	if redirect == "" {
		redirect = "http://127.0.0.1:4382/callback"
	}
	return Config{
		ClientID:     pick(file, envClientID),
		ClientSecret: pick(file, envSecret),
		RedirectURI:  redirect,
		DataDir:      dataDir,
	}, nil
}

func RedactedClientID(id string) string {
	if id == "" {
		return "missing"
	}
	if len(id) <= 4 {
		return "set"
	}
	return id[:2] + "***" + id[len(id)-2:]
}

func pick(file map[string]string, key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return file[key]
}

func loadEnvFile() (map[string]string, error) {
	for _, path := range []string{"mcp/.env", ".env"} {
		values, err := readEnvFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		return values, nil
	}
	return map[string]string{}, nil
}

func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key != "" {
			values[key] = value
		}
	}
	return values, scanner.Err()
}
