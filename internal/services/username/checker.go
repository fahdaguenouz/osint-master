package username

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"osint/internal/core"
	"regexp"
	"strings"
	"time"
)

func checkProfileWithActivity(ctx context.Context, client *http.Client, networkName, url, handle string) (found bool, profileInfo, followers, lastActive string, posts []core.Post, warning string) {
	// Special early handling for TikTok - use OEmbed API
	if networkName == "tiktok" {
		return checkTikTokWithOEmbed(ctx, client, handle)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", "", "", nil, networkName + ": request build failed"
	}

	// Standard headers for other platforms
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
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
		Bio         string `json:"bio"`
		Followers   int    `json:"followers"`
		PublicRepos int    `json:"public_repos"`
		UpdatedAt   string `json:"updated_at"`
		CreatedAt   string `json:"created_at"`
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
	if strings.Contains(html, "page not found") {
		return false, "", "", "", nil, ""
	}
	if strings.Contains(html, "login") && strings.Contains(html, "sign up") {
		return false, "", "", "", nil, "instagram: login wall"
	}

	// 1. Check if Private
	isPrivate := extractField(text, `"is_private":`, `,`)
	if isPrivate == "true" {
		return true, "Account is private", "Hidden", "", nil, "Private Profile"
	}

	// 2. Extract Bio
	bio := extractBetween(text, `<meta property="og:description" content="`, `"`, 300)
	if idx := strings.Index(bio, "•"); idx != -1 {
		bio = strings.TrimSpace(bio[idx+1:]) // Clean up the prefix
	}

	// 3. Extract Followers
	followers := extractField(text, `"edge_followed_by":{"count":`, `}`)

	// 4. Extract Recent Posts
	var posts []core.Post
	// Look inside the edge_owner_to_timeline_media object
	postDates := extractAllMatches(text, `"taken_at_timestamp":(\d+)`)

	for i, ts := range postDates {
		if i >= 3 {
			break
		}
		posts = append(posts, core.Post{
			Content:  fmt.Sprintf("IG Post %d", i+1),
			Date:     unixToDate(ts),
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
	// The syndication API returns a 404 if the user doesn't exist
	if strings.Contains(html, "User not found") {
		return false, "", "", "", nil, ""
	}

	// 1. Extract Bio (from the embedded JSON data)
	bio := extractBetween(text, `"description":"`, `"`, 300)

	// 2. Extract Followers
	followers := extractField(text, `"followers_count":`, `,`)

	// 3. Extract Recent Tweets from the syndication timeline
	var posts []core.Post
	// The syndication API exposes tweet text and created_at nicely
	tweetDates := extractAllMatches(text, `"created_at":"([^"]+)"`)
	tweetTexts := extractAllMatches(text, `"text":"([^"]+)"`)

	for i := 0; i < len(tweetDates) && i < len(tweetTexts); i++ {
		if i >= 3 {
			break
		}

		// Clean up the tweet text slightly
		content := cleanJSONString(tweetTexts[i])
		if len(content) > 50 {
			content = content[:47] + "..."
		}

		posts = append(posts, core.Post{
			Content:  content,
			Date:     parseTwitterDate(tweetDates[i]),
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
