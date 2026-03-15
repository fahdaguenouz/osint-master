package ip

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Provider interface {
	Lookup(ctx context.Context, ip string) (ISP, City string, Lat, Lon float64, Source string, err error)
}

type IPAPIProvider struct {
	client *http.Client
}

func NewIPAPIProvider() *IPAPIProvider {
	return &IPAPIProvider{
		client: &http.Client{Timeout: 6 * time.Second},
	}
}

type ipAPIResp struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Query   string  `json:"query"`
	City    string  `json:"city"`
	ISP     string  `json:"isp"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

func (p *IPAPIProvider) Lookup(ctx context.Context, ip string) (string, string, float64, float64, string, error) {
	// Only request the fields we need (saves bandwidth)
	// docs: https://ip-api.com/docs/api:json
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,query,city,isp,lat,lon", ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", 0, 0, "ip-api.com", err
	}
	req.Header.Set("User-Agent", "osint/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", 0, 0, "ip-api.com", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, 0, "ip-api.com", fmt.Errorf("ip-api.com returned HTTP %d", resp.StatusCode)
	}

	var data ipAPIResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", 0, 0, "ip-api.com", err
	}

	if data.Status != "success" {
		// Example: reserved/private ranges can fail
		if data.Message == "" {
			data.Message = "lookup failed"
		}
		return "", "", 0, 0, "ip-api.com", fmt.Errorf("%s", data.Message)
	}

	return data.ISP, data.City, data.Lat, data.Lon, "ip-api.com", nil
}
