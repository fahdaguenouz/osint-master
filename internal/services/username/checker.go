package username

import (
	"context"
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
		snippet, _ := readSnippet(resp.Body, 2*1024*1024)
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
