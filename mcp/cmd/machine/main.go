package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/machine/tools"
	"shanks/internal/transport"
)

func main() {
	log.SetOutput(os.Stderr)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := applyEnvFile(); err != nil {
		return err
	}
	if os.Getenv("MCP_TRANSPORT") == "" {
		os.Setenv("MCP_TRANSPORT", "http")
	}
	if os.Getenv("MCP_ADDR") == "" {
		os.Setenv("MCP_ADDR", "127.0.0.1:8092")
	}
	log.Printf("machine mcp transport=%s addr=%s", os.Getenv("MCP_TRANSPORT"), os.Getenv("MCP_ADDR"))
	transport.Serve("machine", "0.1.0", func(s *server.MCPServer) {
		tools.Register(s)
	})
	return nil
}

// applyEnvFile loads mcp/machine.env. Values already set in the process win.
func applyEnvFile() error {
	for _, path := range []string{"machine.env", "mcp/machine.env"} {
		values, err := readEnvFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		for key, value := range values {
			if _, ok := os.LookupEnv(key); ok {
				continue
			}
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
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
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			values[key] = value
		}
	}
	return values, scanner.Err()
}
