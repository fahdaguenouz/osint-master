package username

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"osint/internal/core"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func scrapeTwitterPlaywright(page playwright.Page, url, handle string) (bool, string, string, string, []core.Post, string) {
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return false, "", "", "", nil, "twitter: navigation failed"
	}

	title, _ := page.Title()
	if strings.Contains(strings.ToLower(title), "suspended") || strings.Contains(strings.ToLower(title), "doesn't exist") {
		return false, "", "", "", nil, ""
	}

	// Wait for the user profile description div to load (max 5 seconds)
	_, err = page.WaitForSelector("[data-testid='UserDescription']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		// If it times out, the profile likely doesn't exist or is behind a login wall
		return false, "", "", "", nil, ""
	}

	// Extract Bio
	bioText := ""
	if bioLoc := page.Locator("[data-testid='UserDescription']"); bioLoc != nil {
		bioText, _ = bioLoc.InnerText()
	}

	// Extract Followers
	followersText := ""
	// Selects the anchor tag containing the word "followers" in the href
	followLoc := page.Locator("a[href$='/verified_followers'] > div > span > span")
	if count, err := followLoc.InnerText(); err == nil {
		followersText = count
	}

	return true, cleanPlaywrightText(bioText), followersText, "", nil, ""
}

func cleanPlaywrightText(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
}

// --- TWITTER (Syndication API Bypass) ---
func checkTwitterSyndication(ctx context.Context, client *http.Client, handle string) (bool, string, string, string, []core.Post, string) {
	url := fmt.Sprintf("https://cdn.syndication.twimg.com/widgets/followbutton/info.json?screen_names=%s", handle)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, "", "", "", nil, "twitter: request failed"
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, "twitter: connection error"
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, "", "", "", nil, ""
	}

	var data []struct {
		Name           string `json:"name"`
		FollowersCount int    `json:"followers_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || len(data) == 0 {
		return false, "", "", "", nil, ""
	}

	user := data[0]
	followersText := fmt.Sprintf("%d", user.FollowersCount)
	profileInfo := "Name: " + user.Name

	return true, profileInfo, followersText, "", nil, ""
}
