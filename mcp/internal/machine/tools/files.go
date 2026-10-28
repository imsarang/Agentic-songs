package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func grepHandler() server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		root := req.GetString("path", os.Getenv("HOME"))
		out, err := exec.Command("grep", "-R", "-n", "-I", "--", req.GetString("query", ""), root).CombinedOutput()
		if err != nil {
			if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
				return textOrErr(string(out), err)
			}
		}
		text := string(out)
		if len(text) > 4000 {
			text = text[:4000]
		}
		paths := map[string]struct{}{}
		for _, line := range strings.Split(text, "\n") {
			if p, _, ok := strings.Cut(line, ":"); ok && p != "" {
				paths[p] = struct{}{}
			}
		}
		list := make([]string, 0, len(paths))
		for p := range paths {
			list = append(list, p)
		}
		sort.Strings(list)
		return textOrErr("paths="+strings.Join(list, " ")+"\n---\n"+text, nil)
	}
}

func catHandler() server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")
		b, err := os.ReadFile(path)
		if err != nil {
			return textOrErr("", err)
		}
		if bytes.IndexByte(b, 0) >= 0 {
			return textOrErr("", fmt.Errorf("binary file"))
		}
		if len(b) > 8000 {
			b = b[:8000]
		}
		return textOrErr(fmt.Sprintf("path=%s bytes=%d\n---\n%s", path, len(b), b), nil)
	}
}
