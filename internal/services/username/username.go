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
		Timeout: 7 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// do not follow redirects automatically
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	results := make([]core.NetworkResult, 0, len(DefaultNetworks))

	for _, netw := range DefaultNetworks {
		url := netw.URL(handle)

		found, warn := checkProfile(ctx, client, netw.Name, url, handle)
		if warn != "" {
			r.Warnings = append(r.Warnings, warn)
		}

		results = append(results, core.NetworkResult{
			Name:  netw.Name,
			URL:   url,
			Found: found,
		})
	}

	r.Username.Networks = results
	r.Sources = append(r.Sources, "direct HTTP check + HTML fingerprint")

	return r, nil
}
