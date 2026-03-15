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

type SubdomainInfo struct {
    Name        string
    IP          string
    CNAME       string
    SSLValid    bool
    SSLExpiry   string
    TakeoverRisk string // "none", "potential", "confirmed"
}

func Run(query string) (core.Result, error) {
    domain := strings.TrimSpace(query)
    if !isValidDomain(domain) {
        err := core.NewUserError("invalid domain format")
        return core.Fail(core.KindDomain, domain, err), err
    }

    r := core.NewBaseResult(core.KindDomain, domain)
    r.Domain.Domain = domain

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 1. Enumerate subdomains via crt.sh (free, no API key required)
    subdomains, err := enumerateFromCRTSH(ctx, domain)
    if err != nil {
        r.Warnings = append(r.Warnings, fmt.Sprintf("crt.sh enumeration failed: %v", err))
    }

    // 2. Resolve IPs and check SSL for each
    var infos []SubdomainInfo
    for _, sub := range subdomains {
        info := resolveAndCheck(sub)
        infos = append(infos, info)
    }

    // 3. Check for takeover risks
    checkTakeovers(infos)

    r.Domain.Subdomains = infos
    r.Sources = append(r.Sources, "crt.sh", "DNS resolution", "SSL certificate check")

    return r, nil
}

func enumerateFromCRTSH(ctx context.Context, domain string) ([]string, error) {
    // crt.sh is a free certificate transparency log search [^11^]
    url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
    
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("crt.sh returned %d", resp.StatusCode)
    }

    var entries []struct {
        NameValue string `json:"name_value"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
        return nil, err
    }

    // Deduplicate and clean
    seen := make(map[string]bool)
    var subs []string
    for _, e := range entries {
        names := strings.Split(e.NameValue, "\n")
        for _, name := range names {
            name = strings.TrimSpace(name)
            if name == "" || seen[name] {
                continue
            }
            seen[name] = true
            subs = append(subs, name)
        }
    }

    return subs, nil
}

func resolveAndCheck(subdomain string) SubdomainInfo {
    info := SubdomainInfo{Name: subdomain}

    // Get CNAME
    cname, _ := net.LookupCNAME(subdomain)
    info.CNAME = strings.TrimSuffix(cname, ".")

    // Resolve IP
    ips, err := net.LookupHost(subdomain)
    if err == nil && len(ips) > 0 {
        info.IP = ips[0]
    }

    // Check SSL
    if info.IP != "" {
        conf := &tls.Config{
            ServerName: subdomain,
        }
        conn, err := tls.Dial("tcp", info.IP+":443", conf)
        if err == nil {
            defer conn.Close()
            state := conn.ConnectionState()
            if len(state.PeerCertificates) > 0 {
                cert := state.PeerCertificates[0]
                info.SSLValid = true
                info.SSLExpiry = cert.NotAfter.Format("2006-01-02")
            }
        }
    }

    return info
}

// Check for subdomain takeover vulnerabilities [^1^][^7^]
func checkTakeovers(infos []SubdomainInfo) {
    // Known vulnerable service patterns
    takeoverPatterns := map[string]string{
        "s3.amazonaws.com":        "AWS S3",
        "s3-website":              "AWS S3 Website",
        "github.io":               "GitHub Pages",
        "github.com":              "GitHub Pages (legacy)",
        "herokuapp.com":           "Heroku",
        "herokussl.com":           "Heroku SSL",
        "wordpress.com":           "WordPress.com",
        "shopify.com":             "Shopify",
        "fastly":                  "Fastly",
        "cloudfront.net":          "AWS CloudFront",
        "azurewebsites.net":       "Azure Websites",
        "blob.core.windows.net":   "Azure Blob",
        "unbouncepages.com":       "Unbounce",
        "zendesk.com":             "Zendesk",
    }

    for i := range infos {
        info := &infos[i]
        
        // Check if CNAME points to known vulnerable service
        for pattern, service := range takeoverPatterns {
            if strings.Contains(info.CNAME, pattern) {
                // Check if it's actually vulnerable (NXDOMAIN or error)
                if info.IP == "" {
                    info.TakeoverRisk = fmt.Sprintf("Potential takeover: %s (CNAME to %s but no resolution)", service, info.CNAME)
                }
            }
        }

        // Additional check: if subdomain resolves but returns specific error pages
        // This would require HTTP check (simplified here)
    }
}

func isValidDomain(d string) bool {
    // Basic validation - should contain at least one dot and valid chars
    if !strings.Contains(d, ".") {
        return false
    }
    // Remove www. prefix if present for validation
    d = strings.TrimPrefix(d, "www.")
    return len(d) > 3 && len(d) < 253
}