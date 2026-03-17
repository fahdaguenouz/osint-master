package username

import (
	"osint/internal/core"
	"strings"
)

func parseTwitterDetailed(text, html string) (bool, string, string, string, []core.Post, string) {

	if strings.Contains(strings.ToLower(html), "page doesn’t exist") ||
		strings.Contains(strings.ToLower(html), "account suspended") {
		return false, "", "", "", nil, ""
	}

	var bio string
	var followers string
	var posts []core.Post
	var lastActive string

	// ✅ BIO (multiple fallback patterns)
	bioPatterns := []string{
		`"description":"([^"]+)"`,
		`"bio":"([^"]+)"`,
	}

	for _, p := range bioPatterns {
		match := extractAllMatches(text, p)
		if len(match) > 0 {
			bio = cleanJSONString(match[0])
			break
		}
	}

	// ✅ FOLLOWERS (try multiple formats)
	followPatterns := []string{
		`"followers_count":(\d+)`,
		`"followers":\{"count":(\d+)`,
	}

	for _, p := range followPatterns {
		match := extractAllMatches(text, p)
		if len(match) > 0 {
			followers = match[0]
			break
		}
	}

	// ✅ TWEETS
	tweetTexts := extractAllMatches(text, `"text":"([^"]+)"`)
	tweetDates := extractAllMatches(text, `"created_at":"([^"]+)"`)

	for i := 0; i < len(tweetTexts) && i < len(tweetDates); i++ {
		if i >= 3 {
			break
		}

		content := cleanJSONString(tweetTexts[i])
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		date := parseTwitterDate(tweetDates[i])

		posts = append(posts, core.Post{
			Content:  content,
			Date:     date,
			Platform: "Twitter",
		})

		if i == 0 {
			lastActive = date
		}
	}

	return true, bio, followers, lastActive, posts, ""
}