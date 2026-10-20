package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func Register(s *server.MCPServer, bus *runtime.Bus) {
	s.AddTool(mcp.NewTool("spotify_session",
		mcp.WithDescription("Report whether the Spotify API token is ready. No browser."),
	), sessionHandler(bus))

	s.AddTool(mcp.NewTool("spotify_search",
		mcp.WithDescription("Search tracks. Returns handles and does not start playback."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Song or artist text")),
		mcp.WithNumber("limit", mcp.Description("How many rows to return, from 1 to 8")),
	), searchHandler(bus))

	s.AddTool(mcp.NewTool("spotify_play",
		mcp.WithDescription("Play a query's top hit, or a handle from the latest search. Returns the now-playing bar."),
		mcp.WithString("query", mcp.Description("Play the top hit for this text")),
		mcp.WithNumber("search_id", mcp.Description("search_id from the latest spotify_search result")),
		mcp.WithNumber("result_index", mcp.Description("index from that search result")),
		mcp.WithString("title", mcp.Description("title from that search result")),
		mcp.WithString("artist", mcp.Description("artist from that search result")),
	), playHandler(bus))

	s.AddTool(mcp.NewTool("spotify_control",
		mcp.WithDescription("Pause, resume, or skip, then return the now-playing bar."),
		mcp.WithString("action", mcp.Required(), mcp.Description("pause, resume, next, or previous")),
	), controlHandler(bus))

	s.AddTool(mcp.NewTool("spotify_now_playing",
		mcp.WithDescription("Read the player bar. Returns nothing_playing when the bar is empty."),
	), nowPlayingHandler(bus))
}
