package ip

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"osint/src/core"
	"osint/src/detect"
)

func Run(query string) (core.Result, error) {
	q := strings.TrimSpace(query)

	if !detect.IsIPv4(q) {
		err := core.NewUserError("invalid IPv4 address")
		return core.Fail(core.KindIP, q, err), err
	}

	r := core.NewBaseResult(core.KindIP, q)

	// Create context with timeout for all operations
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Initialize providers
	geoProvider := NewIPAPIProvider()
	abuseProvider := NewAbuseIPDBProvider()

	// We use a WaitGroup to run both network requests at the exact same time
	var wg sync.WaitGroup
	wg.Add(2)

	// Variables to hold the results from the goroutines
	var (
		isp, city, country, asn, geoSource string
		lat, lon                           float64
		geoErr                             error

		abuseScore, abuseReports           int
		lastReported, abuseSource          string
		abuseErr                           error
	)

	// 1. Fetch Geolocation Data concurrently
	go func() {
		defer wg.Done()
		isp, city, country, asn, lat, lon, geoSource, geoErr = geoProvider.Lookup(ctx, q)
	}()

	// 2. Fetch Abuse Reputation concurrently
	go func() {
		defer wg.Done()
		abuseScore, abuseReports, lastReported, abuseSource, abuseErr = abuseProvider.CheckIP(ctx, q)
	}()

	// Wait for both API calls to finish
	wg.Wait()

	// Handle GeoIP results
	if geoErr != nil {
		r.Warnings = append(r.Warnings, fmt.Sprintf("Geolocation lookup failed: %v", geoErr))
	} else {
		r.Sources = append(r.Sources, geoSource)
	}

	// Handle AbuseIPDB results
	if abuseErr != nil {
		r.Warnings = append(r.Warnings, fmt.Sprintf("Abuse check skipped: %v", abuseErr))
	} else {
		r.Sources = append(r.Sources, abuseSource)
	}

	// Build the final result
	r.IP = core.IPResult{
		IP:           q,
		ISP:          isp,
		City:         city,
		Country:      country,
		ASN:          asn,
		Lat:          lat,
		Lon:          lon,
		AbuseScore:   abuseScore,
		AbuseReports: abuseReports,
	}

	if abuseErr != nil || abuseScore == 0 {
		r.IP.KnownIssues = "No reported abuse"
	} else {
		r.IP.KnownIssues = fmt.Sprintf("Abuse confidence: %d%% (%d reports)", abuseScore, abuseReports)
		if lastReported != "" {
			r.IP.KnownIssues += fmt.Sprintf(", last reported: %s", lastReported)
		}
	}

	return r, nil
}