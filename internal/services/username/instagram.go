package username

import (
    "fmt"
    "osint/internal/core"
    "strings"

    "github.com/playwright-community/playwright-go"
)

// --- INSTAGRAM (Playwright with fixed Selectors & Stats) ---
func scrapeInstagramPlaywright(page playwright.Page, url, handle string) (bool, string, string, string, []core.Post, string) {
    _, err := page.Goto(url, playwright.PageGotoOptions{
        WaitUntil: playwright.WaitUntilStateDomcontentloaded,
    })
    if err != nil {
        return false, "", "", "", nil, "instagram: navigation failed"
    }

    currentURL := page.URL()
    if strings.Contains(currentURL, "accounts/login") {
        return false, "", "", "", nil, "instagram: forced login wall"
    }

    title, _ := page.Title()
    if strings.Contains(strings.ToLower(title), "page not found") {
        return false, "", "", "", nil, ""
    }

    // FIX: Increased timeout to 10 seconds (10000ms) to allow heavy SPA rendering
    _, err = page.WaitForSelector("header", playwright.PageWaitForSelectorOptions{
        Timeout: playwright.Float(10000), 
    })
    if err != nil {
        return false, "", "", "", nil, "instagram: profile didn't render"
    }

    // 1. Extract Stats (posts, followers, following) from the unordered list
    statsText := ""
    statsLoc := page.Locator("header ul")
    if text, err := statsLoc.InnerText(); err == nil {
        statsText = cleanPlaywrightText(text)
    }

    // 2. Extract Bio from the last div in the header section
    bioText := ""
    bioLoc := page.Locator("header section").Locator("xpath=./div[last()]")
    if text, err := bioLoc.InnerText(); err == nil {
        bioText = cleanPlaywrightText(text)
    }

    // 3. Extract pure follower count
    followersText := ""
    followLoc := page.Locator(fmt.Sprintf("a[href='/%s/followers/'] span", handle)).First()
    if text, err := followLoc.GetAttribute("title"); err == nil && text != "" {
        followersText = text
    } else if text, err := followLoc.InnerText(); err == nil {
        followersText = text
    }

    // 4. Combine Stats and Bio for the final output string
    profileInfo := ""
    if statsText != "" {
        profileInfo += statsText
    }
    
    if bioText != "" {
        if profileInfo != "" {
            profileInfo += " | " // Add a separator if we have both
        }
        profileInfo += "Bio: " + bioText
    }

    return true, profileInfo, followersText, "", nil, ""
}