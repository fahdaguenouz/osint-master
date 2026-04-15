package username

type Network struct {
	Name string
	URL  func(handle string) string
}

// DefaultNetworks - 5 social networks as required
var DefaultNetworks = []Network{
	{Name: "reddit", URL: func(h string) string { return "https://www.reddit.com/user/" + h }},
	{Name: "medium", URL: func(h string) string { return "https://medium.com/@" + h }},
	{Name: "youtube", URL: func(h string) string { return "https://www.youtube.com/@" + h }},
	{Name: "instagram", URL: func(h string) string { return "https://www.instagram.com/" + h + "/" }},
	{Name: "tiktok", URL: func(h string) string { return "https://www.tiktok.com/@" + h }},
	{Name: "github", URL: func(h string) string { return "https://github.com/" + h }},
}