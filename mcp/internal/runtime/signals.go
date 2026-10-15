package runtime

import (
	"os"
	"os/signal"
	"syscall"
)

// Signals waits for SIGINT or SIGTERM, then cancels the root context.
func Signals(cancel func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	cancel()
}
