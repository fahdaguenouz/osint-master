package username

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func extractField(text, start, end string) string {
	idx := strings.Index(text, start)
	if idx == -1 {
		return ""
	}
	idx += len(start)

	// Skip whitespace, quotes, colons
	for idx < len(text) && (text[idx] == ' ' || text[idx] == '"' || text[idx] == ':' || text[idx] == '{') {
		idx++
	}

	endIdx := strings.Index(text[idx:], end)
	if endIdx == -1 {
		endIdx = len(text) - idx
	}
	if endIdx > 200 {
		endIdx = 200
	}

	return strings.TrimSpace(text[idx : idx+endIdx])
}

func extractBetween(text, start, end string, maxLen int) string {
	idx := strings.Index(text, start)
	if idx == -1 {
		return ""
	}
	idx += len(start)

	endIdx := strings.Index(text[idx:], end)
	if endIdx == -1 || endIdx > maxLen {
		endIdx = maxLen
		if endIdx > len(text[idx:]) {
			endIdx = len(text[idx:])
		}
	}

	return strings.TrimSpace(text[idx : idx+endIdx])
}

func extractAllMatches(text, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(text, -1)
	var results []string
	for _, m := range matches {
		if len(m) > 1 {
			results = append(results, m[1])
		}
	}
	return results
}

func cleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	s = strings.ReplaceAll(s, `\n`, " ")
	s = strings.ReplaceAll(s, `\u0026`, "&")
	s = strings.ReplaceAll(s, `\u003c`, "<")
	s = strings.ReplaceAll(s, `\u003e`, ">")
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	if len(s) > 200 {
		s = s[:197] + "..."
	}
	return s
}

func cleanHTML(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	s = re.ReplaceAllString(s, " ")
	return cleanJSONString(s)
}

func unixToDate(ts string) string {
	i := 0
	fmt.Sscanf(ts, "%d", &i)
	if i == 0 {
		return ts
	}
	t := time.Unix(int64(i), 0)
	return t.Format("2006-01-02")
}

func formatGitHubDate(date string) string {
	// GitHub format: 2024-03-15T10:30:00Z
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return t.Format("2006-01-02")
}


func extractRealBio(text string) string {
	lines := strings.Split(text, "\n")

	var clean []string

	for _, l := range lines {
		l = strings.TrimSpace(l)

		// ❌ Skip junk
		if l == "" ||
			strings.Contains(strings.ToLower(l), "followers") ||
			strings.Contains(strings.ToLower(l), "following") ||
			strings.Contains(strings.ToLower(l), "posts") {
			continue
		}

		clean = append(clean, l)
	}

	return strings.Join(clean, " ")
}

func extractNumber(text, key string) string {
	parts := strings.Split(text, ",")
	for _, p := range parts {
		if strings.Contains(strings.ToLower(p), strings.ToLower(key)) {
			return strings.TrimSpace(strings.Split(p, " ")[0])
		}
	}
	return ""
}