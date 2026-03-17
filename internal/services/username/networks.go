package username

type Network struct {
	Name string
	URL  func(handle string) string
}

// DefaultNetworks - 5 social networks as required
var DefaultNetworks = []Network{
	{Name: "facebook", URL: func(h string) string { return "https://www.facebook.com/" + h }},
	{Name: "twitter", URL: func(h string) string { return "https://x.com/" + h }},
	{Name: "instagram", URL: func(h string) string { return "https://www.instagram.com/" + h + "/" }},
	{Name: "tiktok", URL: func(h string) string { return "https://www.tiktok.com/@" + h }},
	{Name: "github", URL: func(h string) string { return "https://github.com/" + h }},
}