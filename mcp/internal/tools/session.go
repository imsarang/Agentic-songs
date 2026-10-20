package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"shanks/internal/runtime"
)

func sessionHandler(bus *runtime.Bus) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return reply(call(ctx, bus, runtime.KindSession, runtime.Args{}))
	}
}

func call(ctx context.Context, bus *runtime.Bus, kind runtime.Kind, args runtime.Args) (runtime.Result, error) {
	cmd := runtime.CommandInit(ctx, kind, args)

	select {
	case bus.CommandsChannel <- cmd:
	case <-ctx.Done():
		return runtime.Result{}, ctx.Err()
	}

	select {
	case result := <-cmd.ReplyChannel:
		return result, nil
	case <-ctx.Done():
		return runtime.Result{}, ctx.Err()
	}
}

// Song tools wait until login has closed ReadyChannel.
func waitReady(ctx context.Context, bus *runtime.Bus) error {
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	select {
	case <-bus.ReadyChannel:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("session_not_ready")
	}
}

func reply(result runtime.Result, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if result.Err != nil {
		return mcp.NewToolResultError(result.Err.Error()), nil
	}
	return mcp.NewToolResultText(result.Text), nil
}
