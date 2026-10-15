package spotify

import (
	"fmt"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func OpenPlayer(page playwright.Page) (View, error) {
	_, err := page.Goto(WebPlayerURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(45000),
	})
	if err != nil {
		return ViewUnknown, err
	}
	if err = dismissCookies(page); err != nil {
		return ViewUnknown, err
	}
	return WaitView(page, 20*time.Second)
}

func WaitView(page playwright.Page, limit time.Duration) (View, error) {
	deadline := time.Now().Add(limit)
	for {
		view, err := Classify(page)
		if err != nil {
			return ViewUnknown, err
		}
		if view != ViewUnknown || time.Now().After(deadline) {
			return view, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func Classify(page playwright.Page) (View, error) {
	login, err := anyVisible(page, `[data-testid="login-button"]`)
	if err != nil {
		return ViewUnknown, err
	}
	if login || roleVisible(page, "Log in") || roleVisible(page, "Continue with Google") {
		return ViewLogin, nil
	}

	// The bottom bar is on screen before login. The account control is the signed-in signal.
	signedIn, err := anyVisible(
		page,
		`[data-testid="user-widget-link"]`,
		`[data-testid="user-widget-avatar"]`,
	)
	if err != nil {
		return ViewUnknown, err
	}
	if signedIn {
		return ViewPlayer, nil
	}
	return ViewUnknown, nil
}

func dismissCookies(page playwright.Page) error {
	for _, name := range []string{"Accept Cookies", "Accept all", "Accept"} {
		ok := roleVisible(page, name)
		if !ok {
			continue
		}
		return clickRole(page, name)
	}
	return nil
}

func anyVisible(page playwright.Page, selectors ...string) (bool, error) {
	for _, selector := range selectors {
		ok, err := page.Locator(selector).First().IsVisible()
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}
	return false, nil
}

func roleVisible(page playwright.Page, name string) bool {
	_, ok := findRole(page, name)
	return ok
}

func clickRole(page playwright.Page, name string) error {
	loc, ok := findRole(page, name)
	if !ok {
		return fmt.Errorf("%s is not on the page", name)
	}
	return loc.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(10000)})
}

func findRole(page playwright.Page, name string) (playwright.Locator, bool) {
	for _, role := range []*playwright.AriaRole{playwright.AriaRoleButton, playwright.AriaRoleLink} {
		loc := page.GetByRole(*role, playwright.PageGetByRoleOptions{Name: name}).First()
		ok, err := loc.IsVisible()
		if err == nil && ok {
			return loc, true
		}
	}
	return nil, false
}
