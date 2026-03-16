package username

import (
	"context"
	"net/http"
	"strings"
	"time"

	"osint/internal/core"
	"osint/internal/detect"
)

func Run(query string) (core.Result, error) {
	q := strings.TrimSpace(query)
	if !detect.IsUsername(q) {
		err := core.NewUserError("invalid username format")
		return core.Fail(core.KindUsername, q, err), err
	}

	handle := strings.TrimPrefix(q, "@")

	r := core.NewBaseResult(core.KindUsername, q)
	r.Username.Username = handle

	client := &http.Client{
		Timeout: 12 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	results := make([]core.NetworkResult, 0, len(DefaultNetworks))
	var activePlatforms []string
	var allPosts []core.Post

	for _, netw := range DefaultNetworks {
		url := netw.URL(handle)

		found, profileInfo, followers, lastActive, posts, warn := checkProfileWithActivity(ctx, client, netw.Name, url, handle)
		if warn != "" {
			r.Warnings = append(r.Warnings, warn)
		}

		if found {
			activePlatforms = append(activePlatforms, netw.Name)
			allPosts = append(allPosts, posts...)
		}

		results = append(results, core.NetworkResult{
			Name:        netw.Name,
			URL:         url,
			Found:       found,
			ProfileInfo: profileInfo,
			Followers:   followers,
			LastActive:  lastActive,
			RecentPosts: posts,
		})
	}

	// Determine most recent activity across all platforms
	r.Username.Networks = results
	r.Username.RecentActivity = formatRecentActivity(activePlatforms)
	
	// Find the newest post
	if len(allPosts) > 0 {
		newest := findNewestPost(allPosts)
		r.Username.LastPost = newest.Content
		r.Username.LastPostDate = newest.Date
		r.Username.LastPostPlatform = newest.Platform
	} else {
		r.Username.LastPost = "No recent public activity found"
	}

	r.Sources = append(r.Sources, "direct HTTP check + HTML fingerprint")

	return r, nil
}

func formatRecentActivity(platforms []string) string {
	if len(platforms) == 0 {
		return "No recent activity detected"
	}
	return "Active on: " + strings.Join(platforms, ", ")
}

func findNewestPost(posts []core.Post) core.Post {
	// Simple string comparison for dates (ISO format: 2024-03-15)
	// In production, you'd parse actual time.Time
	newest := posts[0]
	for _, p := range posts {
		if p.Date > newest.Date {
			newest = p
		}
	}
	return newest
}