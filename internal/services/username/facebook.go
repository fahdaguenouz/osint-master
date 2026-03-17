package username

import (
	"osint/internal/core"
	"strings"
)

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
