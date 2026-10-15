package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const (
	defaultProfileDir = "mcp/data/chrome-profile"
	envEmail          = "SPOTIFY_GOOGLE_EMAIL"
	envPassword       = "SPOTIFY_GOOGLE_PASSWORD"
	envProfile        = "SPOTIFY_PROFILE_DIR"
	envHeadless       = "SPOTIFY_HEADLESS"
)

type Config struct {
	GoogleEmail    string
	GooglePassword string
	ProfileDir     string
	Headless       bool
}

func Load() (Config, error) {
	file, err := loadEnvFile()
	if err != nil {
		return Config{}, err
	}
	profile := pick(file, envProfile)
	if profile == "" {
		profile = defaultProfileDir
	}
	headless, err := parseHeadless(pick(file, envHeadless))
	if err != nil {
		return Config{}, err
	}
	return Config{
		GoogleEmail:    pick(file, envEmail),
		GooglePassword: pick(file, envPassword),
		ProfileDir:     profile,
		Headless:       headless,
	}, nil
}

func RedactedEmail(email string) string {
	if email == "" {
		return "missing"
	}

	at := strings.LastIndex(email, "@")
	if at <= 1 {
		return "set"
	}
	return email[:1] + "***" + email[at:]
}

func pick(file map[string]string, key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return file[key]
}

func parseHeadless(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
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
