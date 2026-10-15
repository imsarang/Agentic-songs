package browser

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
	"shanks/internal/runtime"
	"shanks/internal/spotify"
)

func waitForManualLogin(ctx context.Context, page playwright.Page, profileDir string, publish func(runtime.SessionEvent)) error {
	log.Printf("log in by hand in the open window, including any captcha. the session file will be %s", sessionFile(profileDir))
	publish(runtime.SessionEvent{Type: runtime.EventNeedsLogin, Phase: runtime.PhaseLoggingIn})

	toldRejected := false
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if closedPage(page) {
			return fmt.Errorf("chromium closed during login")
		}
		if !toldRejected && strings.Contains(page.URL(), "signin/rejected") {
			toldRejected = true
			log.Printf("Google blocked the previous sign-in. Use the Spotify login form in the window.")
		}
		view, err := spotify.Classify(page)
		if err != nil {
			return err
		}
		if view == spotify.ViewPlayer {
			return nil
		}
		time.Sleep(time.Second)
	}
}
