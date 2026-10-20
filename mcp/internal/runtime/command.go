package runtime

import "context"

// CommandInit builds a command with a capacity-1 reply mailbox.
func CommandInit(ctx context.Context, kind Kind, args Args) Command {
	if ctx == nil {
		ctx = context.Background()
	}
	return Command{
		Kind:         kind,
		Args:         args,
		ReplyChannel: make(chan Result, 1),
		Ctx:          ctx,
	}
}

func (c Command) Send(result Result) {

	// 1. Guard Clause: If the channel wasn't initialized, exit immediately
	// to prevent a panic or a permanent block.
	if c.ReplyChannel == nil {
		return
	}

	ctx := c.Ctx

	// 2. Fallback: If no context was provided, use a background context
	// so the select statement below doesn't panic on a nil context
	if ctx == nil {
		ctx = context.Background()
	}

	// 3. The Race: Whichever of these two events happens first wins.
	select {
	case c.ReplyChannel <- result:
	case <-ctx.Done():
		// This wins if the context times out or is cancelled BEFORE
		// anyone reads from c.ReplyChannel. The function exits, abandoning the send.
	}
}
