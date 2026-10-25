package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/spotify"
	"shanks/internal/spotify/cache"
)

func Register(s *server.MCPServer, client spotify.Client) {
	found := &cache.Search{}

	s.AddTool(mcp.NewTool("spotify_search",
		mcp.WithDescription("Search tracks. Returns handles and does not start playback."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Song or artist text")),
		mcp.WithNumber("limit", mcp.Description("How many rows to return, from 1 to 8")),
	), searchHandler(client, found))

	s.AddTool(mcp.NewTool("spotify_play",
		mcp.WithDescription("Play a query's top hit, or a handle from the latest search. Returns requested track and the now-playing bar."),
		mcp.WithString("query", mcp.Description("Play the top hit for this text")),
		mcp.WithNumber("search_id", mcp.Description("search_id from the latest spotify_search result")),
		mcp.WithNumber("result_index", mcp.Description("index from that search result")),
		mcp.WithString("title", mcp.Description("title from that search result")),
		mcp.WithString("artist", mcp.Description("artist from that search result")),
	), playHandler(client, found))

	s.AddTool(mcp.NewTool("spotify_control",
		mcp.WithDescription("Pause, resume, or skip, then return the now-playing bar."),
		mcp.WithString("action", mcp.Required(), mcp.Description("pause, resume, next, or previous")),
	), controlHandler(client))

	s.AddTool(mcp.NewTool("spotify_now_playing",
		mcp.WithDescription("Read the player bar. Returns nothing_playing when the bar is empty."),
	), nowPlayingHandler(client))

	s.AddTool(mcp.NewTool("spotify_playlist",
		mcp.WithDescription("List/create/add/play/update a playlist. action=list|create|add|play|update."),
		mcp.WithString("action", mcp.Description("list, create, add, play, or update (default add)")),
		mcp.WithString("playlist_name", mcp.Description("playlist name for create, add, play, or update")),
		mcp.WithString("description", mcp.Description("playlist description for create or update")),
		mcp.WithString("query", mcp.Description("search text; top hit uri is added to the playlist")),
		mcp.WithString("use_current", mcp.Description("true to add the currently playing track")),
		mcp.WithNumber("search_id", mcp.Description("search_id from spotify_search")),
		mcp.WithNumber("result_index", mcp.Description("index from that search")),
		mcp.WithString("title", mcp.Description("title from that search")),
		mcp.WithString("artist", mcp.Description("artist from that search")),
	), playlistHandler(client, found))
}
