package ip

import (
	"context"
	"strings"
	"time"

	"osint/internal/core"
	"osint/internal/detect"
)

func Run(query string) (core.Result, error) {
	q := strings.TrimSpace(query)

	if !detect.IsIPv4(q) {
		err := core.NewUserError("invalid IPv4 address")
		return core.Fail(core.KindIP, q, err), err
	}

	r := core.NewBaseResult(core.KindIP, q)

	// Provider (ip-api.com)
	provider := NewIPAPIProvider()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	isp, city, lat, lon, source, err := provider.Lookup(ctx, q)
	if err != nil {
		// Put the error inside Result for printing/logging,
		// but also return err so main exits with code 1 if you want.
		fail := core.Fail(core.KindIP, q, err)
		fail.Sources = append(fail.Sources, source)
		return fail, err
	}

	r.IP = core.IPResult{
		IP:   q,
		ISP:  isp,
		City: city,
		Lat:  lat,
		Lon:  lon,
	}
	r.Sources = append(r.Sources, source)

	return r, nil
}
