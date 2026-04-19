package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"osint/src/core"
)

func NextResultFilename(dir string) (string, error) {
	if err := ensureDirExists(dir); err != nil {
		return "", err
	}

	base := filepath.Join(dir, "result.txt")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base, nil
	}

	for i := 2; i < 10000; i++ {
		name := filepath.Join(dir, fmt.Sprintf("result%d.txt", i))
		if _, err := os.Stat(name); os.IsNotExist(err) {
			return name, nil
		}
	}
	return "", fmt.Errorf("too many result files in %s", dir)
}

func ensureDirExists(dir string) error {
	if dir == "." {
		return nil
	}
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	return err
}

// WriteResult writes results to file (filename can be custom via -o or auto-generated)
func WriteResult(filename string, r core.Result) error {
	body := formatForFile(r)
	return os.WriteFile(filename, []byte(body), 0o644)
}

func formatForFile(r core.Result) string {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("OSINT-Master Report\n"))
	b.WriteString(fmt.Sprintf("==================\n\n"))
	b.WriteString(fmt.Sprintf("Query: %s\n", r.Input))
	b.WriteString(fmt.Sprintf("Type: %s\n", r.Kind))
	b.WriteString(fmt.Sprintf("Timestamp: %s\n\n", r.Timestamp.Format(time.RFC3339)))

	if r.Error != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n", r.Error))
		return b.String()
	}

	if len(r.Warnings) > 0 {
		b.WriteString("Warnings:\n")
		for _, w := range r.Warnings {
			b.WriteString(fmt.Sprintf("  - %s\n", w))
		}
		b.WriteString("\n")
	}

	// OSINT-Master compatible output format
	switch r.Kind {
	case core.KindDomain:
		b.WriteString(fmt.Sprintf("Main Domain: %s\n\n", r.Input))
		b.WriteString(fmt.Sprintf("Subdomains found: %d\n", len(r.Domain.Subdomains)))

		for _, sub := range r.Domain.Subdomains {
			ip := sub.IP
			if ip == "" {
				ip = "unresolved"
			}
			b.WriteString(fmt.Sprintf("  - %s (IP: %s)\n", sub.Name, ip))
			if sub.CNAME != "" && sub.CNAME != sub.Name {
				b.WriteString(fmt.Sprintf("    CNAME: %s\n", sub.CNAME))
			}

			if sub.SSLValid {
				b.WriteString(fmt.Sprintf("    SSL Certificate: Valid until %s\n", sub.SSLExpiry))
			} else {
				b.WriteString(fmt.Sprintf("    SSL Certificate: Not found\n"))
			}
		}

		// List takeover risks separately
		var riskySubs []core.SubdomainInfo
		for _, sub := range r.Domain.Subdomains {
			if sub.TakeoverRisk != "" && sub.TakeoverRisk != "none" {
				riskySubs = append(riskySubs, sub)
			}
		}

		if len(riskySubs) > 0 {
			b.WriteString("\nPotential Subdomain Takeover Risks:\n")
			for _, sub := range riskySubs {
				b.WriteString(fmt.Sprintf("  - Subdomain: %s\n", sub.Name))
				b.WriteString(fmt.Sprintf("    %s\n", sub.TakeoverRisk))
				b.WriteString("    Recommended Action: Remove or update the DNS record to prevent potential misuse\n")
			}
		} else {
			b.WriteString("\nPotential Subdomain Takeover Risks: None detected\n")
		}

	case core.KindIP:
		// Match required output format exactly per OSINT-Master spec
		if r.IP.ISP != "" {
			b.WriteString(fmt.Sprintf("ISP: %s\n", r.IP.ISP))
		}
		if r.IP.City != "" {
			b.WriteString(fmt.Sprintf("City: %s\n", r.IP.City))
		}
		if r.IP.Country != "" {
			b.WriteString(fmt.Sprintf("Country: %s\n", r.IP.Country))
		}
		if r.IP.ASN != "" {
			b.WriteString(fmt.Sprintf("ASN: %s\n", r.IP.ASN))
		}
		if r.IP.Lat != 0 || r.IP.Lon != 0 {
			b.WriteString(fmt.Sprintf("Lat/Lon: %.4f / %.4f\n", r.IP.Lat, r.IP.Lon))
		}

		// Abuse data / Known Issues
		if r.IP.KnownIssues != "" {
			b.WriteString(fmt.Sprintf("Known Issues: %s\n", r.IP.KnownIssues))
		} else {
			b.WriteString("Known Issues: No reported abuse\n")
		}

	case core.KindUsername:
		// Check presence on at least 5 social networks
		for _, n := range r.Username.Networks {
			val := "Not Found"
			if n.Found {
				val = "Found"
			}

			name := n.Name
			if len(name) > 0 {
				name = strings.ToUpper(name[:1]) + name[1:]
			}

			b.WriteString(fmt.Sprintf("%s: %s", name, val))
			if n.Followers != "" {
				b.WriteString(fmt.Sprintf(" (%s followers)", n.Followers))
			}
			b.WriteString("\n")

			// Profile bio
			if n.ProfileInfo != "" {
				if n.Name == "tiktok" {
					b.WriteString(fmt.Sprintf(" Author: %s\n", n.ProfileInfo))
				} else {
					b.WriteString(fmt.Sprintf("  Bio: %s\n", n.ProfileInfo))
				}
			}

			// Recent posts/activity for this platform
			if len(n.RecentPosts) > 0 {
				b.WriteString(fmt.Sprintf("  Recent Activity:\n"))
				for _, post := range n.RecentPosts {
					b.WriteString(fmt.Sprintf("    - %s", post.Content))
					if post.Date != "" {
						b.WriteString(fmt.Sprintf(" (%s)", post.Date))
					}
					b.WriteString("\n")
				}
			}
		}

		// Summary of recent activity across platforms
		b.WriteString(fmt.Sprintf("\nRecent Activity: %s\n", r.Username.RecentActivity))

		// Most recent post across all platforms
		if r.Username.LastPostPlatform != "" {
			b.WriteString(fmt.Sprintf("Last Post: %s on %s", r.Username.LastPost, r.Username.LastPostPlatform))
			if r.Username.LastPostDate != "" {
				b.WriteString(fmt.Sprintf(" (%s)", r.Username.LastPostDate))
			}
			b.WriteString("\n")
		}

	}

	// Sources used
	if len(r.Sources) > 0 {
		b.WriteString(fmt.Sprintf("\nSources: %s\n", strings.Join(r.Sources, ", ")))
	}

	b.WriteString(fmt.Sprintf("\nData saved in result file\n"))
	return b.String()
}
