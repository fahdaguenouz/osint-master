package core

import "time"

type Kind string

const (
	KindFullName Kind = "full-name"
	KindIP       Kind = "ip"
	KindUsername Kind = "username"
	KindDomain   Kind = "domain"
)

type Result struct {
	Kind      Kind
	Input     string
	Timestamp time.Time
	Sources   []string
	Warnings  []string
	Error     string

	FullName FullNameResult
	IP       IPResult
	Username UsernameResult
	Domain   DomainResult
}

type FullNameResult struct {
	FirstName string
	LastName  string
	Address   string
	Phone     string
}

type IPResult struct {
	IP           string
	ISP          string
	City         string
	Country      string
	ASN          string
	Lat          float64
	Lon          float64
	AbuseScore   int
	AbuseReports int
	KnownIssues  string
}

type UsernameResult struct {
	Username       string
	Networks       []NetworkResult
	RecentActivity string // Summary of recent activity across platforms
	LastPost       string // The most recent post/activity found
	LastPostDate   string // When it was posted
	LastPostPlatform string // Which platform had the newest activity
}

type NetworkResult struct {
	Name        string
	URL         string
	Found       bool
	ProfileInfo string // Bio/description
	Followers   string // Follower count
	LastActive  string // Last activity date if available
	RecentPosts []Post // Recent posts/activity (up to 3)
}

type Post struct {
	Content   string
	Date      string
	Platform  string
	URL       string
}

type DomainResult struct {
	Domain     string
	Subdomains []SubdomainInfo
}

type SubdomainInfo struct {
	Name         string
	IP           string
	CNAME        string
	SSLValid     bool
	SSLExpiry    string
	TakeoverRisk string
}

func NewBaseResult(kind Kind, input string) Result {
	return Result{
		Kind:      kind,
		Input:     input,
		Timestamp: time.Now(),
	}
}

func Fail(kind Kind, input string, err error) Result {
	r := NewBaseResult(kind, input)
	if err != nil {
		r.Error = err.Error()
	}
	return r
}