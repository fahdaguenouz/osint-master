package username

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"osint/internal/core"
	"strings"
)

// checkTikTokWithOEmbed - Single function for TikTok checking using OEmbed API
func checkTikTokWithOEmbed(ctx context.Context, client *http.Client, handle string) (bool, string, string, string, []core.Post, string) {
	oembedURL := fmt.Sprintf("https://www.tiktok.com/oembed?url=https://www.tiktok.com/@%s", handle)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, oembedURL, nil)
	if err != nil {
		return false, "", "", "", nil, "tiktok: request failed"
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, "tiktok: connection failed"
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return checkTikTokDirect(ctx, client, handle)
	}

	var data struct {
		AuthorName string `json:"author_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.AuthorName == "" {
		return checkTikTokDirect(ctx, client, handle)
	}
	
	profileInfo :=  data.AuthorName
	// Build profile info from author_name only

	// Get followers and posts from profile page
	followers, posts := fetchTikTokStats(ctx, client, handle)

	return true, profileInfo, followers, "", posts, ""
}

// fetchTikTokStats - Fetches only stats (followers, posts) from TikTok profile page
func fetchTikTokStats(ctx context.Context, client *http.Client, handle string) (string, []core.Post) {
	url := fmt.Sprintf("https://www.tiktok.com/@%s", handle)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://www.google.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", nil
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", nil
		}
		defer gzReader.Close()
		reader = gzReader
	}

	snippet, _ := io.ReadAll(io.LimitReader(reader, 256*1024))
	text := string(snippet)

	// Extract followers only
	followers := extractField(text, `"followerCount":`, `,`)
	if followers == "" {
		followers = extractField(text, `"fans":`, `,`)
	}
	if followers == "" {
		followers = extractField(text, `"stats":{"followerCount":`, `}`)
	}

	// Extract video dates (max 3)
	var posts []core.Post
	videoDates := extractAllMatches(text, `"createTime":"(\d{10})"`)
	for i, ts := range videoDates {
		if i >= 3 {
			break
		}
		posts = append(posts, core.Post{
			Content:  fmt.Sprintf("Video %d", i+1),
			Date:     unixToDate(ts),
			Platform: "TikTok",
		})
	}

	return followers, posts
}

// checkTikTokDirect - Fallback direct check if OEmbed fails
func checkTikTokDirect(ctx context.Context, client *http.Client, handle string) (bool, string, string, string, []core.Post, string) {
	url := fmt.Sprintf("https://www.tiktok.com/@%s", handle)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, "", "", "", nil, "tiktok: failed"
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", "", nil, "tiktok: connection error"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, "", "", "", nil, ""
	}

	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	text := string(snippet)
	html := strings.ToLower(text)

	notFoundPatterns := []string{
		"couldn't find this account",
		"couldn&#39;t find this account",
		"user not found",
		"account not found",
	}
	for _, pattern := range notFoundPatterns {
		if strings.Contains(html, pattern) {
			return false, "", "", "", nil, ""
		}
	}

	if strings.Contains(text, `"uniqueId":"`+handle+`"`) ||
		strings.Contains(text, `"uniqueId": "`+handle+`"`) ||
		strings.Contains(text, `userInfo`) {
		followers, posts := fetchTikTokStats(ctx, client, handle)
		return true, "", followers, "", posts, "tiktok: limited data"
	}

	return false, "", "", "", nil, "tiktok: profile not found"
}
