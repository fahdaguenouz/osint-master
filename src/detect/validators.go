package detect

import "regexp"

var (
	reIPv4 = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|1?\d?\d)$`)
	// very simple: two words, letters + hyphen/apostrophe allowed
	reFullName = regexp.MustCompile(`^[A-Za-zÀ-ÖØ-öø-ÿ'’-]+(?:\s+)[A-Za-zÀ-ÖØ-öø-ÿ'’-]+$`)
	// username like "@user01" or "user01"
	reUsername = regexp.MustCompile(`^@?[a-zA-Z0-9._-]{2,32}$`)
)

func IsIPv4(s string) bool     { return reIPv4.MatchString(s) }
func IsFullName(s string) bool { return reFullName.MatchString(s) }
func IsUsername(s string) bool { return reUsername.MatchString(s) }
