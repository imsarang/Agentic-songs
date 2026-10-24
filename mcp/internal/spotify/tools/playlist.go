package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/spotify"
	"shanks/internal/spotify/cache"
)

func playlistHandler(client spotify.Client, found *cache.Search) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.EnsureAuth(ctx); err != nil {
			return textOrErr("", err)
		}
		action := strings.TrimSpace(req.GetString("action", "add"))
		name := strings.TrimSpace(req.GetString("playlist_name", ""))
		if name == "" && action != "list" {
			return mcp.NewToolResultError("playlist_name required"), nil
		}
		desc := strings.TrimSpace(req.GetString("description", ""))
		useCurrent := req.GetString("use_current", "")
		wantCurrent := useCurrent == "true" || useCurrent == "1"

		switch action {
		case "list":
			list, err := client.ListPlaylists(ctx)
			if err != nil {
				return textOrErr("", err)
			}
			return textOrErr(spotify.FormatPlaylists(list), nil)
		case "create":
			pl, err := client.CreatePlaylist(ctx, name, desc)
			if err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("id=%s name=%s uri=%s", pl.ID, pl.Name, pl.URI), nil)
		case "add":
			pl, err := spotify.FindPlaylist(ctx, client, name)
			if err != nil {
				return textOrErr("", err)
			}
			uri, err := resolveTrackURI(ctx, client, found, wantCurrent, req)
			if err != nil {
				return textOrErr("", err)
			}
			if err := client.AddToPlaylist(ctx, pl.ID, uri); err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("added uri=%s to playlist=%s id=%s", uri, pl.Name, pl.ID), nil)
		case "play":
			pl, err := spotify.FindPlaylist(ctx, client, name)
			if err != nil {
				return textOrErr("", err)
			}
			uri := pl.URI
			if uri == "" {
				uri = "spotify:playlist:" + pl.ID
			}
			if err := client.PlayPlaylist(ctx, uri); err != nil {
				return textOrErr("", err)
			}
			bar, err := client.NowPlaying(ctx)
			if err != nil {
				return textOrErr("", err)
			}
			return textOrErr(spotify.FormatBar(bar), nil)
		case "update":
			pl, err := spotify.FindPlaylist(ctx, client, name)
			if err != nil {
				return textOrErr("", err)
			}
			if err := client.UpdatePlaylistMetadata(ctx, pl.ID, pl.Name, desc); err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("updated id=%s name=%s description=%s", pl.ID, pl.Name, desc), nil)
		case "delete":
			pl, err := spotify.FindPlaylist(ctx, client, name)
			if err != nil {
				return textOrErr("", err)
			}
			if err := client.DeletePlaylist(ctx, pl.ID); err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("deleted(unfollowed) id=%s name=%s", pl.ID, pl.Name), nil)
		case "remove":
			pl, err := spotify.FindPlaylist(ctx, client, name)
			if err != nil {
				return textOrErr("", err)
			}
			uri, err := resolveTrackURI(ctx, client, found, wantCurrent, req)
			if err != nil {
				return textOrErr("", err)
			}
			if err := client.RemoveFromPlaylist(ctx, pl.ID, uri); err != nil {
				return textOrErr("", err)
			}
			return textOrErr(fmt.Sprintf("removed uri=%s from playlist=%s id=%s", uri, pl.Name, pl.ID), nil)
		default:
			return mcp.NewToolResultError("action must be list, create, add, play, or update"), nil
		}
	}
}

func resolveTrackURI(ctx context.Context, client spotify.Client, found *cache.Search, useCurrent bool, req mcp.CallToolRequest) (string, error) {
	if useCurrent {
		bar, err := client.NowPlaying(ctx)
		if err != nil {
			return "", err
		}
		if bar.URI == "" {
			return "", fmt.Errorf("nothing_playing")
		}
		return bar.URI, nil
	}
	hit, err := found.Pick(
		ctx,
		client,
		strings.TrimSpace(req.GetString("query", "")),
		req.GetInt("search_id", 0),
		req.GetInt("result_index", 0),
		strings.TrimSpace(req.GetString("title", "")),
	)
	if err != nil {
		return "", err
	}
	return hit.URI, nil
}
