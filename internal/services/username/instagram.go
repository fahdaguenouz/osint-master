package username

import (
	"osint/internal/core"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// --- INSTAGRAM (Improved Anti-Detection + Fallbacks) ---
func scrapeInstagramPlaywright(page playwright.Page, url, handle string) (bool, string, string, string, []core.Post, string) {

	// -------------------------------
	// 1. Stealth Patch (IMPORTANT)
	// -------------------------------
	page.AddInitScript(playwright.Script{
		Content: playwright.String(`
		Object.defineProperty(navigator, 'webdriver', {
			get: () => false,
		});
	`),
	})
	// -------------------------------
	// 2. Navigate (NO networkidle)
	// -------------------------------
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(15000),
	})
	if err != nil {
		return false, "", "", "", nil, "instagram: navigation failed"
	}

	// Give React time to render
	page.WaitForTimeout(3000)

	currentURL := page.URL()

	// -------------------------------
	// 3. Detect Login Wall
	// -------------------------------
	if strings.Contains(currentURL, "/accounts/login") {
		return true, "Profile exists but redirected to login", "", "", nil, "instagram: login redirect"
	}
	// -------------------------------
	// 4. Detect Not Found
	// -------------------------------
	title, _ := page.Title()
	if strings.Contains(strings.ToLower(title), "not found") {
		return false, "", "", "", nil, ""
	}

	// -------------------------------
	// 5. Wait for profile header
	// -------------------------------
	_, err = page.WaitForSelector("header", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(8000),
	})

	// Even if header fails → don't exit yet (fallback later)

	var profileInfo string
	var followers string

	// -------------------------------
	// 6. STRATEGY 1: META TAG (Best case)
	// -------------------------------
	metaDesc, _ := page.Locator(`meta[name="description"]`).GetAttribute("content")

	if metaDesc == "" {
		return true, "Profile may exist but Instagram blocked data", "", "", nil, "instagram: partial block"
	}

	if metaDesc != "" && strings.Contains(metaDesc, "Followers") {
		parts := strings.Split(metaDesc, " - ")
		if len(parts) > 0 {
			statsRaw := parts[0]
			profileInfo = statsRaw

			fParts := strings.Split(statsRaw, " Followers")
			if len(fParts) > 0 {
				followers = strings.TrimSpace(fParts[0])
			}
		}
	}

	// -------------------------------
	// 7. STRATEGY 2: DOM fallback
	// -------------------------------
	if profileInfo == "" {
		if ulText, err := page.Locator("header ul").InnerText(); err == nil {
			profileInfo = strings.ReplaceAll(strings.TrimSpace(ulText), "\n", ", ")
		}
	}

	// -------------------------------
	// 8. Extract BIO
	// -------------------------------
	bioText := ""

	bioLoc := page.Locator("header section > div").Last()
	if text, err := bioLoc.InnerText(); err == nil {
		bioText = cleanPlaywrightText(text)
	}

	if bioText != "" {
		if profileInfo != "" {
			profileInfo += " | Bio: "
		} else {
			profileInfo = "Bio: "
		}
		profileInfo += bioText
	}

	// -------------------------------
	// 9. FINAL FALLBACK (important)
	// -------------------------------
	if profileInfo == "" {
		// profile likely exists but blocked partially
		return true, "Profile exists (limited data - possible blocking)", "", "", nil, "instagram: partial data"
	}

	return true, profileInfo, followers, "", nil, ""
}
