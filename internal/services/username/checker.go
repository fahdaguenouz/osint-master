package username

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"osint/internal/core"
)

func checkProfileWithActivity(ctx context.Context, client *http.Client, networkName, url, handle string) (found bool, profileInfo, followers, lastActive string, posts []core.Post, warning string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", "", "", nil, networkName + ": request build failed"
	}

	// Rotate user agents
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
	req.Header.Set("User-Agent", userAgents[0])
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, networkName + ": request failed"
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	loc := strings.ToLower(resp.Header.Get("Location"))

	// Handle blocks/rate limits
	if code == 401 || code == 403 || code == 429 || code == 999 {
		return false, "", "", "", nil, networkName + ": blocked/rate limited"
	}

	// Handle redirects
	if code == 301 || code == 302 || code == 303 || code == 307 || code == 308 {
		if strings.Contains(loc, "login") || strings.Contains(loc, "signin") {
			return false, "", "", "", nil, networkName + ": login required"
		}
		if networkName == "github" && strings.Contains(loc, "github.com/"+strings.ToLower(handle)) {
			// GitHub canonical redirect - profile exists, fetch with trailing slash
			return fetchGitHubWithRepos(ctx, client, handle)
		}
		return false, "", "", "", nil, networkName + ": redirected"
	}

	if code == 404 || code == 410 {
		return false, "", "", "", nil, ""
	}

	if code == 200 {
		// Read larger snippet for better parsing
		snippet, _ := readSnippet(resp.Body, 512*1024)
		html := strings.ToLower(snippet)
		text := snippet

		switch networkName {
		case "github":
			return parseGitHubWithRepos(text, html, handle, client)
		case "instagram":
			return parseInstagramDetailed(text, html)
		case "twitter":
			return parseTwitterDetailed(text, html)
		case "facebook":
			return parseFacebookDetailed(text, html)
		case "tiktok":
			return parseTikTokDetailed(text, html)
		}
	}

	return false, "", "", "", nil, networkName + ": unexpected status " + fmt.Sprintf("%d", code)
}

// fetchGitHubWithRepos fetches profile + public repos via GitHub API (no auth needed for public)
func fetchGitHubWithRepos(ctx context.Context, client *http.Client, handle string) (bool, string, string, string, []core.Post, string) {
	// First fetch user profile
	profileURL := fmt.Sprintf("https://api.github.com/users/%s", handle)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return false, "", "", "", nil, ""
	}
	req.Header.Set("User-Agent", "osintmaster/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, "", "", "", nil, ""
	}

	var user struct {
		Bio       string `json:"bio"`
		Followers int    `json:"followers"`
		PublicRepos int `json:"public_repos"`
		UpdatedAt string `json:"updated_at"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return false, "", "", "", nil, ""
	}

	// Fetch recent repos (up to 4)
	reposURL := fmt.Sprintf("https://api.github.com/users/%s/repos?sort=updated&per_page=4", handle)
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, reposURL, nil)
	req2.Header.Set("User-Agent", "osintmaster/1.0")
	
	resp2, err := client.Do(req2)
	if err != nil {
		// Return basic profile even if repos fail
		followersStr := fmt.Sprintf("%d", user.Followers)
		return true, user.Bio, followersStr, user.UpdatedAt, nil, ""
	}
	defer resp2.Body.Close()

	var repos []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		UpdatedAt   string `json:"updated_at"`
		Language    string `json:"language"`
	}
	json.NewDecoder(resp2.Body).Decode(&repos)

	// Build posts from repos
	var posts []core.Post
	for _, repo := range repos {
		desc := repo.Description
		if desc == "" {
			desc = "Repository: " + repo.Name
		}
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		posts = append(posts, core.Post{
			Content:  desc,
			Date:     formatGitHubDate(repo.UpdatedAt),
			Platform: "GitHub",
			URL:      fmt.Sprintf("https://github.com/%s/%s", handle, repo.Name),
		})
	}

	followersStr := fmt.Sprintf("%d", user.Followers)
	return true, user.Bio, followersStr, formatGitHubDate(user.UpdatedAt), posts, ""
}

// parseGitHubWithRepos uses HTML parsing as fallback
// parseGitHubWithRepos uses HTML parsing as fallback
func parseGitHubWithRepos(text, html, handle string, client *http.Client) (bool, string, string, string, []core.Post, string) {
	if strings.Contains(html, "page not found") || strings.Contains(html, "404") {
		return false, "", "", "", nil, ""
	}

	// Try API first for better data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Quick API check
	apiFound, bio, followers, lastActive, posts, _ := fetchGitHubWithRepos(ctx, client, handle)
	if apiFound && len(posts) > 0 {
		return apiFound, bio, followers, lastActive, posts, ""
	}

	// Fallback to HTML parsing
	bio = extractField(text, `"bio":`, `,`)
	if bio == "" {
		bio = extractBetween(text, `<div class="p-note user-profile-bio mb-3 js-user-profile-bio f4">`, `</div>`, 300)
		bio = cleanHTML(bio)
	}

	followers = extractField(text, `"followers":`, `,`)
	if followers == "" {
		followers = extractField(text, `"followersCount":`, `,`)
	}

	// Look for repo names in HTML - FIX: use = not := to avoid redeclaration
	var htmlPosts []core.Post
	repoMatches := extractAllMatches(text, `<a href="/`+handle+`/([^"]+)" itemprop="name codeRepository"`)
	for i, repo := range repoMatches {
		if i >= 4 {
			break
		}
		htmlPosts = append(htmlPosts, core.Post{
			Content:  "Repository: " + repo,
			Date:     "", // Unknown from HTML
			Platform: "GitHub",
		})
	}

	return true, cleanJSONString(bio), cleanJSONString(followers), "", htmlPosts, ""
}
// parseInstagramDetailed extracts bio, followers, and last post date
func parseInstagramDetailed(text, html string) (bool, string, string, string, []core.Post, string) {
	// Not found check
	if strings.Contains(html, "page not found") || 
	   strings.Contains(html, "sorry, this page isn't available") ||
	   strings.Contains(html, "the link you followed may be broken") ||
	   strings.Contains(html, "content unavailable") {
		return false, "", "", "", nil, ""
	}

	// Login wall check
	if (strings.Contains(html, "log in") || strings.Contains(html, "login")) && 
	   (strings.Contains(html, "sign up") || strings.Contains(html, "signup")) {
		return false, "", "", "", nil, "instagram: login wall"
	}

	// Extract bio from sharedData or meta
	bio := extractField(text, `"biography":`, `,`)
	if bio == "" {
		bio = extractBetween(text, `<meta property="og:description" content="`, `"`, 300)
		// Remove "X (@handle) • Instagram photos and videos" prefix
		if idx := strings.Index(bio, "•"); idx != -1 {
			bio = strings.TrimSpace(bio[idx+1:])
		}
	}

	// Extract followers
	followers := extractField(text, `"edge_followed_by":{"count":`, `}`)
	if followers == "" {
		followers = extractField(text, `"followers_count":`, `,`)
	}

	// Extract recent post dates (up to 3)
	var posts []core.Post
	postDates := extractAllMatches(text, `"taken_at_timestamp":(\d+)`)
	for i, ts := range postDates {
		if i >= 3 {
			break
		}
		date := unixToDate(ts)
		posts = append(posts, core.Post{
			Content:  fmt.Sprintf("Post %d", i+1),
			Date:     date,
			Platform: "Instagram",
		})
	}

	lastActive := ""
	if len(posts) > 0 {
		lastActive = posts[0].Date
	}

	return true, cleanJSONString(bio), cleanJSONString(followers), lastActive, posts, ""
}

// parseTwitterDetailed extracts bio, followers, recent tweets
func parseTwitterDetailed(text, html string) (bool, string, string, string, []core.Post, string) {
	if strings.Contains(html, "this account doesn’t exist") ||
	   strings.Contains(html, "this account doesn't exist") ||
	   strings.Contains(html, "account suspended") ||
	   strings.Contains(html, "page not found") {
		return false, "", "", "", nil, ""
	}

	if strings.Contains(html, "sign in to x") || 
	   (strings.Contains(html, "log in") && strings.Contains(html, "x.com")) {
		return false, "", "", "", nil, "twitter: login required"
	}

	// Extract bio
	bio := extractBetween(text, `<meta property="og:description" content="`, `"`, 300)
	if bio == "" {
		bio = extractField(text, `"description":`, `,`)
	}

	// Extract followers
	followers := extractField(text, `"followers_count":`, `,`)
	if followers == "" {
		followers = extractField(text, `"followersCount":`, `,`)
	}

	// Extract recent tweet dates
	var posts []core.Post
	tweetDates := extractAllMatches(text, `"created_at":"([^"]+)"`)
	for i, date := range tweetDates {
		if i >= 3 {
			break
		}
		posts = append(posts, core.Post{
			Content:  fmt.Sprintf("Tweet %d", i+1),
			Date:     parseTwitterDate(date),
			Platform: "Twitter",
		})
	}

	return true, cleanJSONString(bio), cleanJSONString(followers), "", posts, ""
}

// parseFacebookDetailed tries to extract basic info
func parseFacebookDetailed(text, html string) (bool, string, string, string, []core.Post, string) {
	// Not available check
	if strings.Contains(html, "this page isn't available") ||
	   strings.Contains(html, "page may have been removed") ||
	   strings.Contains(html, "content isn't available") ||
	   strings.Contains(html, "page not found") {
		return false, "", "", "", nil, ""
	}

	// Login check
	if strings.Contains(html, "log into facebook") ||
	   strings.Contains(html, "log in to continue") ||
	   strings.Contains(html, "login required") {
		return false, "", "", "", nil, "facebook: login required"
	}

	// Try to extract basic info from meta
	bio := extractBetween(text, `<meta name="description" content="`, `"`, 300)
	
	// Try to find any public posts (rare without login)
	var posts []core.Post
	
	// Look for post timestamps in the HTML (very rare to work)
	postMatches := extractAllMatches(text, `data-utime="(\d+)"`)
	for i, ts := range postMatches {
		if i >= 2 {
			break
		}
		posts = append(posts, core.Post{
			Content:  "Public post",
			Date:     unixToDate(ts),
			Platform: "Facebook",
		})
	}

	return true, cleanJSONString(bio), "", "", posts, ""
}

// parseTikTokDetailed with improved detection
func parseTikTokDetailed(text, html string) (bool, string, string, string, []core.Post, string) {
	// Multiple "not found" patterns
	notFoundPatterns := []string{
		"couldn't find this account",
		"couldn&#39;t find this account",
		"couldn’t find this account",
		"page not available",
		"content not available",
		"user not found",
		"account not found",
	}

	for _, pattern := range notFoundPatterns {
		if strings.Contains(html, pattern) {
			return false, "", "", "", nil, ""
		}
	}

	// Check for verification/captcha walls
	if strings.Contains(html, "/captcha") ||
	   strings.Contains(html, "recaptcha") ||
	   strings.Contains(html, "verify to continue") ||
	   strings.Contains(html, "security verification") ||
	   strings.Contains(html, "verify you're a human") {
		return false, "", "", "", nil, "tiktok: verification required"
	}

	// Check for login wall
	if (strings.Contains(html, "log in") || strings.Contains(html, "login")) &&
	   (strings.Contains(html, "sign up") || strings.Contains(html, "for you")) {
		// This might be a login wall, but let's check if we have profile data anyway
		// Sometimes TikTok shows some data even with login prompt
	}

	// Extract signature/bio - try multiple patterns
	bio := extractField(text, `"signature":`, `,`)
	if bio == "" {
		bio = extractField(text, `"desc":`, `,`)
	}
	if bio == "" {
		bio = extractBetween(text, `<h2 class="share-desc">`, `</h2>`, 200)
	}
	if bio == "" {
		// Try meta description
		bio = extractBetween(text, `<meta property="og:description" content="`, `"`, 300)
	}

	// Extract followers - try multiple patterns
	followers := extractField(text, `"followerCount":`, `,`)
	if followers == "" {
		followers = extractField(text, `"fans":`, `,`)
	}
	if followers == "" {
		followers = extractField(text, `"stats":{"followerCount":`, `}`)
	}

	// Extract recent video dates (up to 4)
	var posts []core.Post
	videoDates := extractAllMatches(text, `"createTime":"(\d+)"`)
	if len(videoDates) == 0 {
		videoDates = extractAllMatches(text, `"createTime":(\d+)`)
	}
	
	for i, ts := range videoDates {
		if i >= 4 {
			break
		}
		posts = append(posts, core.Post{
			Content:  fmt.Sprintf("Video %d", i+1),
			Date:     unixToDate(ts),
			Platform: "TikTok",
		})
	}

	// If we found videos or bio, consider it found
	if len(posts) > 0 || bio != "" || followers != "" {
		return true, cleanJSONString(bio), cleanJSONString(followers), "", posts, ""
	}

	// If we see typical TikTok profile HTML structure but no data, might be rate limited
	if strings.Contains(html, "tiktok") && strings.Contains(html, "user") {
		return false, "", "", "", nil, "tiktok: rate limited or data restricted"
	}

	return false, "", "", "", nil, ""
}

// Helper functions
func extractField(text, start, end string) string {
	idx := strings.Index(text, start)
	if idx == -1 {
		return ""
	}
	idx += len(start)

	// Skip whitespace, quotes, colons
	for idx < len(text) && (text[idx] == ' ' || text[idx] == '"' || text[idx] == ':' || text[idx] == '{') {
		idx++
	}

	endIdx := strings.Index(text[idx:], end)
	if endIdx == -1 {
		endIdx = len(text) - idx
	}
	if endIdx > 200 {
		endIdx = 200
	}

	return strings.TrimSpace(text[idx : idx+endIdx])
}

func extractBetween(text, start, end string, maxLen int) string {
	idx := strings.Index(text, start)
	if idx == -1 {
		return ""
	}
	idx += len(start)

	endIdx := strings.Index(text[idx:], end)
	if endIdx == -1 || endIdx > maxLen {
		endIdx = maxLen
		if endIdx > len(text[idx:]) {
			endIdx = len(text[idx:])
		}
	}

	return strings.TrimSpace(text[idx : idx+endIdx])
}

func extractAllMatches(text, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(text, -1)
	var results []string
	for _, m := range matches {
		if len(m) > 1 {
			results = append(results, m[1])
		}
	}
	return results
}

func cleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	s = strings.ReplaceAll(s, `\n`, " ")
	s = strings.ReplaceAll(s, `\u0026`, "&")
	s = strings.ReplaceAll(s, `\u003c`, "<")
	s = strings.ReplaceAll(s, `\u003e`, ">")
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	if len(s) > 200 {
		s = s[:197] + "..."
	}
	return s
}

func cleanHTML(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	s = re.ReplaceAllString(s, " ")
	return cleanJSONString(s)
}

func unixToDate(ts string) string {
	i := 0
	fmt.Sscanf(ts, "%d", &i)
	if i == 0 {
		return ts
	}
	t := time.Unix(int64(i), 0)
	return t.Format("2006-01-02")
}

func parseTwitterDate(date string) string {
	t, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", date)
	if err != nil {
		return date
	}
	return t.Format("2006-01-02")
}

func formatGitHubDate(date string) string {
	// GitHub format: 2024-03-15T10:30:00Z
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return t.Format("2006-01-02")
}

func readSnippet(r io.Reader, max int64) (string, error) {
	b, err := io.ReadAll(io.LimitReader(r, max))
	if err != nil {
		return "", err
	}
	return string(b), nil
}