package username

import (
	"osint/internal/core"
	"regexp"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// --- MEDIUM (Developer-Friendly Platform - No Auth Wall) ---
func scrapeMediumPlaywright(page playwright.Page, url, handle string) (bool, string, string, string, []core.Post, string) {

	// Basic stealth
	page.AddInitScript(playwright.Script{
		Content: playwright.String(`
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
			delete window.__playwright;
		`),
	})

	page.SetExtraHTTPHeaders(map[string]string{
		"Accept-Language": "en-US,en;q=0.9",
	})

	// Navigate
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(15000),
	})
	if err != nil {
		return false, "", "", "", nil, "medium: navigation failed"
	}

	page.WaitForTimeout(3000)

	// Check for not found
	title, _ := page.Title()
	if strings.Contains(title, "404") || 
	   strings.Contains(title, "Not Found") ||
	   strings.Contains(title, "Page not found") ||
	   strings.Contains(strings.ToLower(title), "medium member") {
		return false, "", "", "", nil, ""
	}

	// Check if redirected to home (profile doesn't exist)
	currentURL := page.URL()
	if currentURL == "https://medium.com/" || 
	   currentURL == "https://medium.com" ||
	   strings.Contains(currentURL, "/search?") {
		return false, "", "", "", nil, ""
	}

	// Extract profile info
	var displayName, bioText, followers string

	// Strategy 1: Meta tags (Medium has excellent meta tags)
	metaTitle, _ := page.Locator(`meta[property="og:title"]`).GetAttribute("content")
	if metaTitle != "" {
		displayName = strings.TrimSpace(metaTitle)
		// Clean up suffixes
		if idx := strings.LastIndex(displayName, " – Medium"); idx != -1 {
			displayName = strings.TrimSpace(displayName[:idx])
		}
	}

	metaDesc, _ := page.Locator(`meta[property="og:description"]`).GetAttribute("content")
	if metaDesc != "" {
		bioText = strings.TrimSpace(metaDesc)
	}

	// Strategy 2: Extract from page content
	pageText, _ := page.Locator(`body`).InnerText()
	if pageText != "" {
		// Look for follower count: "X Followers"
		followersRegex := regexp.MustCompile(`(?i)([\d,]+(?:\.\d+)?[KMBkmb]?)\s+followers?`)
		if match := followersRegex.FindStringSubmatch(pageText); len(match) >= 2 {
			followers = match[1]
		}
	}

	// Strategy 3: Try specific selectors for Medium
	if displayName == "" {
		nameSelectors := []string{
			`h1`,
			`[data-testid="authorName"]`,
			`.author-name`,
			`[rel="author"]`,
		}
		for _, sel := range nameSelectors {
			if text, err := page.Locator(sel).First().InnerText(); err == nil && text != "" {
				displayName = strings.TrimSpace(text)
				break
			}
		}
	}

	// Strategy 4: Extract from title
	if displayName == "" {
		// Title format: "Name – Medium"
		if idx := strings.LastIndex(title, " – Medium"); idx != -1 {
			displayName = strings.TrimSpace(title[:idx])
		} else if idx := strings.LastIndex(title, " | Medium"); idx != -1 {
			displayName = strings.TrimSpace(title[:idx])
		}
	}

	if displayName == "" {
		return false, "", "", "", nil, "medium: unable to extract profile data"
	}

	// Build profile info
	profileInfo := displayName
	if bioText != "" {
		if len(bioText) > 200 {
			bioText = bioText[:200] + "..."
		}
		profileInfo += " | " + bioText
	}

	return true, profileInfo, followers, "", nil, ""
}
