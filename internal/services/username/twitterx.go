package username

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"osint/internal/core"
	"strings"
)



type twitterSyndicationUser struct {
	Name                    string `json:"name"`
	ScreenName              string `json:"screen_name"`
	FollowersCount          int    `json:"followers_count"`
	FormattedFollowersCount string `json:"formatted_followers_count"`
	Protected               bool   `json:"protected"`
	Verified                bool   `json:"verified"`
}

func checkTwitterSyndication(ctx context.Context, client *http.Client, handle string) (bool, string, string, string, []core.Post, string) {
	url := fmt.Sprintf("https://cdn.syndication.twimg.com/widgets/followbutton/info.json?screen_names=%s", handle)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", "", "", nil, "twitter: request creation failed"
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json,text/plain,*/*")
	req.Header.Set("Referer", "https://x.com/")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, "twitter: connection error"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, "", "", "", nil, ""
	}
	if resp.StatusCode != http.StatusOK {
		return false, "", "", "", nil, fmt.Sprintf("twitter: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", "", "", nil, "twitter: read failed"
	}

	if len(strings.TrimSpace(string(body))) == 0 {
		return false, "", "", "", nil, "twitter: empty response"
	}

	var data []twitterSyndicationUser
	if err := json.Unmarshal(body, &data); err != nil {
		return false, "", "", "", nil, "twitter: invalid json"
	}

	if len(data) == 0 || data[0].ScreenName == "" {
		return false, "", "", "", nil, ""
	}

	user := data[0]

	profileInfo := fmt.Sprintf("Name: %s | Username: @%s", user.Name, user.ScreenName)
	if user.Verified {
		profileInfo += " | Verified"
	}
	if user.Protected {
		profileInfo += " | Protected"
	}

	followers := fmt.Sprintf("%d", user.FollowersCount)

	lastActive := ""
	if user.FormattedFollowersCount != "" {
		lastActive = user.FormattedFollowersCount
	}

	return true, profileInfo, followers, lastActive, nil, ""
}