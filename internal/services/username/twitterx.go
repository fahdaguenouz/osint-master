package username

import (
	"osint/internal/core"
	"strings"
)

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
