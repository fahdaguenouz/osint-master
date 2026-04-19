package domain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"osint/src/core"
)

// ===================== GLOBALS =====================

var cache = make(map[string][]string)
var cacheMutex sync.RWMutex

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	"Mozilla/5.0 (X11; Linux x86_64)",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
}

var client = &http.Client{
	Timeout: 25 * time.Second,
	Transport: &http.Transport{
		DisableKeepAlives: true,
	},
}

// ===================== MAIN =====================

func Run(query string) (core.Result, error) {
	domain := cleanDomain(query)

	if !isValidDomain(domain) {
		err := core.NewUserError("invalid domain")
		return core.Fail(core.KindDomain, domain, err), err
	}

	r := core.NewBaseResult(core.KindDomain, domain)
	r.Domain.Domain = domain

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// CACHE
	cacheMutex.RLock()
	if cached, ok := cache[domain]; ok {
		cacheMutex.RUnlock()
		return buildResult(ctx, r, cached), nil
	}
	cacheMutex.RUnlock()

	var allSubs []string

	// ===== SOURCE 1: CRT.SH =====
	crtSubs, err := crtshWithRetry(ctx, domain)
	if err != nil {
		r.Warnings = append(r.Warnings, "crt.sh failed")
	}
	allSubs = append(allSubs, crtSubs...)

	// ===== SOURCE 2: BUFFEROVER =====
	bufSubs, err := bufferover(ctx, domain)
	if err != nil {
		r.Warnings = append(r.Warnings, "bufferover failed")
	}
	allSubs = append(allSubs, bufSubs...)

	// ===== SOURCE 3: HACKERTARGET =====
	hackSubs, err := hackertarget(ctx, domain)
	if err != nil {
		r.Warnings = append(r.Warnings, "hackertarget failed")
	}
	allSubs = append(allSubs, hackSubs...)

	// ===== SOURCE 4: BRUTE =====
	allSubs = append(allSubs, bruteForce(domain)...)

	allSubs = deduplicate(allSubs, domain)

	if len(allSubs) > 10 {
		cacheMutex.Lock()
		cache[domain] = allSubs
		cacheMutex.Unlock()
	}

	return buildResult(ctx, r, allSubs), nil
}

// ===================== CRT.SH =====================

func crtshWithRetry(ctx context.Context, domain string) ([]string, error) {
	var lastErr error

	for i := 0; i < 3; i++ {
		subs, err := crtsh(ctx, domain)
		if err == nil && len(subs) > 0 {
			return subs, nil
		}
		lastErr = err
		time.Sleep(time.Duration(4+i*2) * time.Second)
	}
	return nil, lastErr
}

func crtsh(ctx context.Context, domain string) ([]string, error) {
	time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)

	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data []struct {
		NameValue string `json:"name_value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var subs []string

	for _, d := range data {
		for _, name := range strings.Split(d.NameValue, "\n") {
			name = strings.TrimSpace(strings.ToLower(name))
			name = strings.TrimPrefix(name, "*.")

			if strings.Contains(name, domain) && isValidSubdomain(name) {
				subs = append(subs, name)
			}
		}
	}

	return subs, nil
}

// ===================== BUFFEROVER =====================

func bufferover(ctx context.Context, domain string) ([]string, error) {
	url := fmt.Sprintf("https://dns.bufferover.run/dns?q=.%s", domain)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var raw map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var subs []string

	parse := func(entries []string) {
		for _, entry := range entries {
			parts := strings.Split(entry, ",")
			if len(parts) == 2 {
				host := strings.TrimSpace(parts[1])
				if strings.HasSuffix(host, domain) {
					subs = append(subs, host)
				}
			}
		}
	}

	if v, ok := raw["FDNS_A"]; ok {
		parse(v)
	}
	if v, ok := raw["RDNS"]; ok {
		parse(v)
	}

	if len(subs) == 0 {
		return nil, fmt.Errorf("no data")
	}

	return subs, nil
}

// ===================== HACKERTARGET =====================

func hackertarget(ctx context.Context, domain string) ([]string, error) {
	url := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")

	var subs []string
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			host := strings.TrimSpace(parts[0])
			if strings.HasSuffix(host, domain) {
				subs = append(subs, host)
			}
		}
	}

	if len(subs) == 0 {
		return nil, fmt.Errorf("no results")
	}

	return subs, nil
}

// ===================== BRUTE =====================

func bruteForce(domain string) []string {
	wordlist := []string{
		"www", "mail", "api", "dev", "test", "admin", "beta",
		"staging", "prod", "app", "portal", "dashboard",
		"auth", "login", "cdn", "static", "img", "files",
	}

	var results []string
	for _, w := range wordlist {
		results = append(results, w+"."+domain)
	}
	return results
}

// ===================== BUILD RESULT =====================

func buildResult(ctx context.Context, r core.Result, subs []string) core.Result {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 15)
	out := make(chan core.SubdomainInfo, len(subs))

	for _, sub := range subs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			sem <- struct{}{}
			out <- analyze(ctx, s)
			<-sem
		}(sub)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	for i := range out {
		r.Domain.Subdomains = append(r.Domain.Subdomains, i)
	}

	r.Sources = []string{"crt.sh", "bufferover", "hackertarget", "dns brute-force"}
	return r
}

// ===================== ANALYZE =====================

func analyze(ctx context.Context, sub string) core.SubdomainInfo {
	info := core.SubdomainInfo{Name: sub, TakeoverRisk: "none"}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ip, _ := net.DefaultResolver.LookupHost(ctx, sub)
	if len(ip) > 0 {
		info.IP = ip[0]
		checkSSL(ctx, sub, ip[0], &info)
	}

	cname, _ := net.DefaultResolver.LookupCNAME(ctx, sub)
	info.CNAME = strings.TrimSuffix(cname, ".")

	return info
}

// ===================== SSL =====================

func checkSSL(ctx context.Context, sub, ip string, info *core.SubdomainInfo) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "443"), 3*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         sub,
		InsecureSkipVerify: true,
	})

	if err := tlsConn.Handshake(); err == nil {
		state := tlsConn.ConnectionState()
		if len(state.PeerCertificates) > 0 {
			info.SSLValid = true
			info.SSLExpiry = state.PeerCertificates[0].NotAfter.Format("2006-01-02")
		}
	}
}

// ===================== HELPERS =====================

func deduplicate(list []string, root string) []string {
	seen := map[string]struct{}{root: {}}
	res := []string{root}

	for _, v := range list {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}

func cleanDomain(d string) string {
	d = strings.ToLower(strings.TrimSpace(d))
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "www.")
	d = strings.TrimSuffix(d, "/")
	return d
}

func isValidDomain(d string) bool {
	return strings.Contains(d, ".")
}

func isValidSubdomain(s string) bool {
	if strings.Contains(s, "@") {
		return false
	}
	if strings.Contains(s, " ") {
		return false
	}
	return strings.Count(s, ".") >= 2
}