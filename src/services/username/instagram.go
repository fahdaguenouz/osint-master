package username

import (
	"osint/src/core"
	"regexp"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// --- INSTAGRAM (Debug Version) ---
func scrapeInstagramPlaywright(page playwright.Page, url, handle string) (bool, string, string, string, []core.Post, string) {

	// ... (stealth and navigation code stays the same) ...

	page.AddInitScript(playwright.Script{
		Content: playwright.String(`
		Object.defineProperty(navigator, 'webdriver', { get: () => false });
		delete window.__playwright;
		delete window.__pw_manual;
	`),
	})

	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(15000),
	})
	if err != nil {
		return false, "", "", "", nil, "instagram: navigation failed"
	}

	page.WaitForTimeout(3000)
	currentURL := page.URL()

	if strings.Contains(currentURL, "/accounts/login") {
		return true, "Profile exists but redirected to login", "", "", nil, "instagram: login redirect"
	}

	title, _ := page.Title()
	if strings.Contains(strings.ToLower(title), "not found") {
		return false, "", "", "", nil, ""
	}

	page.WaitForSelector("header", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(8000),
	})

	// -------------------------------
	// EXTRACT ALL FROM META DESCRIPTION
	// -------------------------------
	var followers, following, posts, bioText string

	metaDesc, _ := page.Locator(`meta[name="description"]`).GetAttribute("content")

	// DEBUG: Print what we got
	// fmt.Printf("DEBUG - Meta description: %q\n", metaDesc)

	if metaDesc != "" {
		// Use SEPARATE regexes for each stat (more reliable)
		// Followers
		followersRegex := regexp.MustCompile(`(?i)([\d,]+)\s+followers?`)
		if match := followersRegex.FindStringSubmatch(metaDesc); len(match) >= 2 {
			followers = strings.TrimSpace(match[1])
			// fmt.Printf("DEBUG - Found followers: %s\n", followers)
		}

		// Following - CRITICAL: use \b word boundary
		followingRegex := regexp.MustCompile(`(?i)([\d,]+)\s+following\b`)
		if match := followingRegex.FindStringSubmatch(metaDesc); len(match) >= 2 {
			following = strings.TrimSpace(match[1])
			// fmt.Printf("DEBUG - Found following: %s\n", following)
		}

		// Posts
		postsRegex := regexp.MustCompile(`(?i)([\d,]+)\s+posts?\b`)
		if match := postsRegex.FindStringSubmatch(metaDesc); len(match) >= 2 {
			posts = strings.TrimSpace(match[1])
			// fmt.Printf("DEBUG - Found posts: %s\n", posts)
		}

		// Extract bio
		postsIdx := strings.Index(metaDesc, "Posts -")
		if postsIdx != -1 {
			bioStart := postsIdx + 7
			seeInstaIdx := strings.LastIndex(metaDesc, " - See Instagram")
			if seeInstaIdx != -1 && seeInstaIdx > bioStart {
				bioText = strings.TrimSpace(metaDesc[bioStart:seeInstaIdx])
			} else {
				bioText = strings.TrimSpace(metaDesc[bioStart:])
			}
			// fmt.Printf("DEBUG - Found bio: %q\n", bioText)
		}
	}

	// ... (rest of the code stays the same) ...

	// Build output
	var infoParts []string
	if followers != "" {
		infoParts = append(infoParts, followers+" followers")
	}
	if following != "" {
		infoParts = append(infoParts, following+" following")
	}
	if posts != "" {
		infoParts = append(infoParts, posts+" posts")
	}

	statsLine := strings.Join(infoParts, ", ")
	finalProfile := statsLine
	if bioText != "" {
		if finalProfile != "" {
			finalProfile += " | " + bioText
		} else {
			finalProfile = bioText
		}
	}

	if finalProfile == "" {
		return true, "Profile exists (limited data)", "", "", nil, "instagram: partial data"
	}

	return true, finalProfile, followers, "", nil, ""
}

func cleanText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	return text
}
