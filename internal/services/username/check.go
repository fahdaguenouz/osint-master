package username

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func checkProfile(ctx context.Context, client *http.Client, networkName, url, handle string) (bool, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, networkName + ": request build failed"
	}

	// Reduce download
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) osint/1.0")
	req.Header.Set("Accept", "text/html,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return false, networkName + ": request failed"
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	loc := strings.ToLower(resp.Header.Get("Location"))

	// If blocked/rate-limited, we can't confirm
	if code == 401 || code == 403 || code == 429 || code == 999 {
		return false, networkName + ": blocked/rate limited (cannot confirm)"
	}

	// Handle redirects (don’t auto-follow)
	if code == 301 || code == 302 || code == 303 || code == 307 || code == 308 {
		// If redirecting to login, can't confirm
		if strings.Contains(loc, "login") || strings.Contains(loc, "signin") || strings.Contains(loc, "auth") {
			return false, networkName + ": redirected to login (cannot confirm)"
		}

		// GitHub sometimes redirects to canonical profile URL (slash)
		if networkName == "github" {
			if strings.Contains(loc, "github.com/"+strings.ToLower(handle)) {
				return true, ""
			}
		}

		// For most other networks, redirects are ambiguous → mark as unknown
		return false, networkName + ": redirected (cannot confirm)"
	}

	// Clear not-found
	if code == 404 || code == 410 {
		return false, ""
	}

	// Many sites return 200 even when not found → fingerprint HTML
	if code == 200 {
		snippet, _ := readSnippet(resp.Body, 64*1024) // 64KB
		html := strings.ToLower(snippet)

		switch networkName {
		case "github":
			// GitHub usually 404 when not found, but just in case:
			if strings.Contains(html, "not found") && strings.Contains(html, "404") {
				return false, ""
			}
			return true, ""

		case "instagram":
			// Common “not found” markers
			if strings.Contains(html, "sorry, this page isn't available") ||
				strings.Contains(html, "page isn't available") {
				return false, ""
			}
			// Login wall sometimes returns 200
			if strings.Contains(html, "log in") && strings.Contains(html, "password") {
				return false, "instagram: login wall (cannot confirm)"
			}
			return true, ""

		case "facebook":
			// Facebook returns 200 for lots of “not found” pages
			if strings.Contains(html, "this page isn't available") ||
				strings.Contains(html, "page may have been removed") ||
				strings.Contains(html, "content isn't available right now") {
				return false, ""
			}
			if strings.Contains(html, "log in") && strings.Contains(html, "facebook") {
				return false, "facebook: login wall (cannot confirm)"
			}
			return true, ""

		case "twitter":
			// X often shows “account doesn’t exist” with 200
			if strings.Contains(html, "this account doesn’t exist") ||
				strings.Contains(html, "this account doesn't exist") ||
				strings.Contains(html, "try searching for another") {
				return false, ""
			}
			if strings.Contains(html, "log in") && strings.Contains(html, "x") {
				return false, "twitter: login wall (cannot confirm)"
			}
			return true, ""

		case "tiktok":
			// TikTok: be careful, HTML may contain the word "captcha" in scripts even when no captcha is shown.
			// We only mark as captcha wall if we see strong markers.
			if strings.Contains(html, "couldn't find this account") ||
				strings.Contains(html, "couldn&#39;t find this account") ||
				strings.Contains(html, "couldn’t find this account") {
				return false, ""
			}

			// Strong verification/captcha markers (much safer than just "captcha")
			if strings.Contains(html, "/captcha") ||
				strings.Contains(html, "recaptcha") ||
				strings.Contains(html, "hcaptcha") ||
				strings.Contains(html, "verify to continue") ||
				strings.Contains(html, "security verification") ||
				strings.Contains(html, "verify you're a human") ||
				strings.Contains(html, "verification required") {
				return false, "tiktok: verification wall (cannot confirm)"
			}

			// Generic not available
			if strings.Contains(html, "page isn't available") || strings.Contains(html, "page is not available") {
				return false, ""
			}

			// If we got a normal 200 and no strong "not found" markers, assume found.
			return true, ""

		default:
			// generic: if it's 200 and no known "not found" marker, assume found
			return true, ""
		}
	}

	// Anything else: unknown -> false with warning
	return false, networkName + ": unexpected HTTP status (cannot confirm)"
}

func readSnippet(r io.Reader, max int64) (string, error) {
	b, err := io.ReadAll(io.LimitReader(r, max))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
