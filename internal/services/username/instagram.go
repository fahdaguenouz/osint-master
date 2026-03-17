package username

import (
	"fmt"
	"osint/internal/core"
	"strings"
)

func parseInstagramDetailed(text, html string) (bool, string, string, string, []core.Post, string) {

	if strings.Contains(html, "page not found") {
		return false, "", "", "", nil, ""
	}

	// 🔥 better login wall detection
	if strings.Contains(html, "login") ||
		strings.Contains(html, "sign up") ||
		strings.Contains(html, "accounts/login") {
		return false, "", "", "", nil, "instagram: login wall"
	}

	var bio string
	var followers string
	var posts []core.Post
	var lastActive string

	// ✅ NEW: try multiple patterns (Instagram changed structure)

	// BIO
	bioPatterns := []string{
		`"biography":"([^"]*)"`,
		`"bio":"([^"]*)"`,
	}

	for _, p := range bioPatterns {
		match := extractAllMatches(text, p)
		if len(match) > 0 {
			bio = cleanJSONString(match[0])
			break
		}
	}

	// FOLLOWERS
	followPatterns := []string{
		`"edge_followed_by":\{"count":(\d+)`,
		`"follower_count":(\d+)`,
	}

	for _, p := range followPatterns {
		match := extractAllMatches(text, p)
		if len(match) > 0 {
			followers = match[0]
			break
		}
	}

	// POSTS
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

		if i == 0 {
			lastActive = date
		}
	}

	// PRIVATE ACCOUNT
	if strings.Contains(text, `"is_private":true`) {
		return true, "Private account", "Hidden", "", nil, "Private Profile"
	}

	return true, bio, followers, lastActive, posts, ""
}
