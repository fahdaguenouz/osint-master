package username

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"osint/internal/core"
)

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
