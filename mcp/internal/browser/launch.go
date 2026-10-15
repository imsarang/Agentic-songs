package browser

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mxschmitt/playwright-go"
	"shanks/internal/config"
)

type launched struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	ctx     playwright.BrowserContext
	page    playwright.Page
}

func launch(cfg *config.Config) (*launched, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	out := &launched{pw: pw}

	opts := playwright.BrowserTypeLaunchOptions{
		Headless:          playwright.Bool(cfg.Headless),
		Channel:           playwright.String("chrome"),
		Args:              []string{"--disable-blink-features=AutomationControlled", "--start-maximized"},
		IgnoreDefaultArgs: []string{"--enable-automation"},
	}
	out.browser, err = pw.Chromium.Launch(opts)
	if err != nil {
		log.Printf("installed Chrome did not start (%v); opening Playwright Chromium", err)
		opts.Channel = nil
		out.browser, err = pw.Chromium.Launch(opts)
	}
	if err != nil {
		out.close()
		return nil, err
	}

	contextOpts := playwright.BrowserNewContextOptions{NoViewport: playwright.Bool(true)}
	if authFile := sessionFile(cfg.ProfileDir); fileExists(authFile) {
		contextOpts.StorageStatePath = playwright.String(authFile)
		log.Printf("loading session file %s", authFile)
	}
	out.ctx, err = out.browser.NewContext(contextOpts)
	if err != nil {
		out.close()
		return nil, err
	}
	out.page, err = out.ctx.NewPage()
	if err != nil {
		out.close()
		return nil, err
	}
	out.page.SetDefaultTimeout(10000)
	return out, nil
}

func sessionFile(profileDir string) string {
	return filepath.Join(filepath.Dir(profileDir), "spotify-auth.json")
}

func saveSession(page playwright.Page, profileDir string) error {
	path := sessionFile(profileDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if _, err := page.Context().StorageState(playwright.BrowserContextStorageStateOptions{
		Path:      playwright.String(path),
		IndexedDB: playwright.Bool(true),
	}); err != nil {
		return err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	log.Printf("session file written to %s", path)
	return nil
}

func (s *launched) close() {
	if s == nil {
		return
	}
	if s.ctx != nil {
		_ = s.ctx.Close()
	}
	if s.browser != nil {
		_ = s.browser.Close()
	}
	if s.pw != nil {
		_ = s.pw.Stop()
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func closedPage(page playwright.Page) bool {
	return page == nil || page.IsClosed()
}

func browserClosed(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "closed")
}
