package username

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"osint/internal/core"
	"strings"
	"time"
)

// RedditUser represents the data structure from /about.json
type RedditUser struct {
	Kind string `json:"kind"`
	Data struct {
		Name              string  `json:"name"`
		DisplayName       string  `json:"display_name"`
		PublicDescription string  `json:"public_description"`
		TotalKarma        int     `json:"total_karma"`
		LinkKarma         int     `json:"link_karma"`
		CommentKarma      int     `json:"comment_karma"`
		CreatedUTC        float64 `json:"created_utc"`
		// Add more fields as needed
	} `json:"data"`
}

// --- REDDIT (Proper JSON API with correct User-Agent) ---
func checkRedditJSON(ctx context.Context, client *http.Client, handle string) (bool, string, string, string, []core.Post, string) {
	handle = strings.ToLower(handle)
	url := fmt.Sprintf("https://www.reddit.com/user/%s/about.json", handle)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, "", "", "", nil, "reddit: request creation failed"
	}

	// CRITICAL: Reddit requires a valid Reddit-style User-Agent
	req.Header.Set("User-Agent", "osint-master/1.0 (by /u/fahdaguenouz)")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, "reddit: connection failed"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, "", "", "", nil, ""
	}

	// Check for rate limit or error
	if resp.StatusCode != 200 {
		return false, "", "", "", nil, fmt.Sprintf("reddit: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", "", "", nil, "reddit: read failed"
	}

	var redditResp RedditUser
	if err := json.Unmarshal(body, &redditResp); err != nil {
		return parseRedditFallback(string(body), handle)
	}

	user := redditResp.Data
	if user.Name == "" {
		return false, "", "", "", nil, ""
	}

	info := fmt.Sprintf(
		"Author: %s | Bio: %s | Karma: %d (Links: %d, Comments: %d)",
		user.Name, // or user.DisplayName if you prefer
		cleanJSONString(user.PublicDescription),
		user.TotalKarma,
		user.LinkKarma,
		user.CommentKarma,
	)

	lastActive := fmtDate(user.CreatedUTC)
	followers := fmt.Sprintf("%d karma", user.TotalKarma)

	return true, info, followers, lastActive, nil, ""
}

// Fallback if JSON structure changed
func parseRedditFallback(data, handle string) (bool, string, string, string, []core.Post, string) {
	if strings.Contains(data, `"error": 404`) || !strings.Contains(data, `"kind": "t2"`) {
		return false, "", "", "", nil, ""
	}

	name := extractBetween(data, `"name":"`, `"`, 100)
	bio := extractBetween(data, `"public_description":"`, `"`, 200)
	karma := extractBetween(data, `"total_karma":`, `,`, 20)

	info := ""
	if name != "" {
		info += "Author: " + cleanJSONString(name)
	}
	if bio != "" {
		info += " | Bio: " + cleanJSONString(bio)
	}
	if karma != "" {
		info += " | Karma: " + karma
	}

	return true, info, "", "", nil, ""
}

func fmtDate(createdUTC float64) string {
	t := time.Unix(int64(createdUTC), 0)
	return t.Format("2006-01-02")
}
