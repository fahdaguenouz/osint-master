package domain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"osint/internal/core"
)

// Run executes domain enumeration and takeover detection
func Run(query string) (core.Result, error) {
	domain := strings.TrimSpace(query)
	if !isValidDomain(domain) {
		err := core.NewUserError("invalid domain format (expected: example.com)")
		return core.Fail(core.KindDomain, domain, err), err
	}

	r := core.NewBaseResult(core.KindDomain, domain)
	r.Domain.Domain = domain

	// Overall timeout for entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 1. Enumerate subdomains via crt.sh (Certificate Transparency logs)
	subdomains, err := enumerateFromCRTSH(ctx, domain)
	if err != nil {
		r.Warnings = append(r.Warnings, fmt.Sprintf("crt.sh enumeration failed: %v", err))
		// Fallback: at least check the main domain
		subdomains = []string{domain}
	}

	// Deduplicate and limit to prevent excessive scanning
	subdomains = deduplicateAndLimit(subdomains, domain, 50)

	// 2. Analyze each subdomain concurrently with timeouts
	var infos []core.SubdomainInfo
	infoChan := make(chan core.SubdomainInfo, len(subdomains))

	// Process with limited concurrency to avoid overwhelming the system
	semaphore := make(chan struct{}, 10) // Max 10 concurrent checks

	for _, sub := range subdomains {
		go func(subdomain string) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			info := analyzeSubdomainWithTimeout(subdomain, 5*time.Second)
			infoChan <- info
		}(sub)
	}

	// Collect results with overall timeout
	collected := 0
	for {
		select {
		case info := <-infoChan:
			infos = append(infos, info)
			collected++
			if collected >= len(subdomains) {
				goto Done
			}
		case <-ctx.Done():
			r.Warnings = append(r.Warnings, "Timeout reached, some subdomains may not have been analyzed")
			goto Done
		}
	}

Done:
	r.Domain.Subdomains = infos
	r.Sources = []string{"crt.sh", "DNS resolution", "SSL certificate check", "CNAME analysis"}

	return r, nil
}

// enumerateFromCRTSH queries crt.sh for certificate transparency logs
func enumerateFromCRTSH(ctx context.Context, domain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "osintmaster/1.0")

	client := &http.Client{
		Timeout: 25 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh returned HTTP %d", resp.StatusCode)
	}

	var entries []struct {
		NameValue string `json:"name_value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var subs []string
	for _, e := range entries {
		names := strings.Split(e.NameValue, "\n")
		for _, name := range names {
			name = strings.TrimSpace(strings.ToLower(name))
			name = strings.TrimPrefix(name, "*.")
			if name == "" || seen[name] {
				continue
			}
			if strings.HasSuffix(name, domain) || name == domain {
				seen[name] = true
				subs = append(subs, name)
			}
		}
	}

	return subs, nil
}

// deduplicateAndLimit cleans list and limits to max items
func deduplicateAndLimit(subs []string, rootDomain string, max int) []string {
	seen := make(map[string]bool)
	var result []string

	// Always include root domain first
	checks := append([]string{rootDomain}, subs...)

	for _, sub := range checks {
		sub = strings.TrimSpace(strings.ToLower(sub))
		if sub == "" || seen[sub] {
			continue
		}
		seen[sub] = true
		result = append(result, sub)
		if len(result) >= max {
			break
		}
	}
	return result
}

// analyzeSubdomainWithTimeout performs analysis with a hard timeout
func analyzeSubdomainWithTimeout(subdomain string, timeout time.Duration) core.SubdomainInfo {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	info := core.SubdomainInfo{
		Name:         subdomain,
		TakeoverRisk: "none",
	}

	// Create a channel for the result
	type result struct {
		ip    string
		cname string
	}
	resChan := make(chan result, 1)

	go func() {
		var res result
		// Get CNAME with timeout via resolver
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 2 * time.Second}
				return d.DialContext(ctx, network, address)
			},
		}

		cname, _ := r.LookupCNAME(ctx, subdomain)
		res.cname = strings.TrimSuffix(cname, ".")

		// Resolve IP
		ips, _ := r.LookupHost(ctx, subdomain)
		if len(ips) > 0 {
			res.ip = ips[0]
		}

		resChan <- res
	}()

	// Wait for DNS resolution or timeout
	select {
	case res := <-resChan:
		info.IP = res.ip
		info.CNAME = res.cname
	case <-ctx.Done():
		// Timeout on DNS, continue with empty data
	}

	// Check SSL only if we have an IP (with short timeout)
	if info.IP != "" {
		checkSSLWithTimeout(subdomain, info.IP, &info, 3*time.Second)
	}

	// Check for takeover if we have a CNAME
	if info.CNAME != "" && info.CNAME != subdomain {
		checkTakeover(info.CNAME, &info)
	}

	return info
}

// checkSSLWithTimeout attempts SSL connection with timeout
func checkSSLWithTimeout(subdomain, ip string, info *core.SubdomainInfo, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use a dialer with timeout
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, "443"))
	if err != nil {
		return
	}
	defer conn.Close()

	// Perform TLS handshake
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         subdomain,
		InsecureSkipVerify: true,
	})
	
	// Set deadline for handshake
	tlsConn.SetDeadline(time.Now().Add(timeout))
	
	if err := tlsConn.Handshake(); err != nil {
		return
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		info.SSLValid = true
		info.SSLExpiry = cert.NotAfter.Format("2006-01-02")
	}
}

// checkTakeover detects dangling CNAME records pointing to vulnerable services
func checkTakeover(cname string, info *core.SubdomainInfo) {
	takeoverPatterns := map[string]string{
		"s3.amazonaws.com":      "AWS S3 bucket",
		"s3-website":            "AWS S3 website",
		"github.io":             "GitHub Pages",
		"github.com":            "GitHub Pages (legacy)",
		"herokuapp.com":         "Heroku",
		"herokussl.com":         "Heroku SSL",
		"wordpress.com":         "WordPress.com",
		"shopify.com":           "Shopify",
		"fastly.net":            "Fastly",
		"fastly.com":            "Fastly",
		"cloudfront.net":        "AWS CloudFront",
		"azurewebsites.net":     "Azure Websites",
		"azurestaticapps.net":   "Azure Static Apps",
		"blob.core.windows.net": "Azure Blob Storage",
		"unbouncepages.com":     "Unbounce",
		"zendesk.com":           "Zendesk",
		"bitbucket.io":          "Bitbucket",
		"ghost.io":              "Ghost",
		"firebaseapp.com":       "Firebase",
		"web.app":               "Firebase Hosting",
		"surge.sh":              "Surge.sh",
		"netlify.app":           "Netlify",
		"vercel.app":            "Vercel",
		"pages.dev":             "Cloudflare Pages",
		"readthedocs.io":        "ReadTheDocs",
		"statuspage.io":         "Statuspage",
	}

	for pattern, service := range takeoverPatterns {
		if strings.Contains(cname, pattern) {
			// Quick check if CNAME target resolves
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, err := net.DefaultResolver.LookupHost(ctx, cname)
			cancel()

			if err != nil {
				info.TakeoverRisk = fmt.Sprintf("CNAME record points to a non-existent %s (%s)", service, cname)
			}
			break
		}
	}
}

// isValidDomain performs basic domain validation
func isValidDomain(d string) bool {
	// Remove protocol if present
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "www.")
	d = strings.TrimSuffix(d, "/")
	d = strings.TrimSpace(d)

	if d == "" || !strings.Contains(d, ".") {
		return false
	}

	// Allow subdomains like sub.example.com - just check basic structure
	if strings.HasPrefix(d, ".") || strings.HasSuffix(d, ".") {
		return false
	}

	return len(d) >= 4 && len(d) <= 253
}