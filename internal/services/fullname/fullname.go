package fullname

import (
	"strings"
	"time"

	"osint/internal/core"
	"osint/internal/detect"
)

func Run(query string) (core.Result, error) {
	q := strings.TrimSpace(query)
	if !detect.IsFullName(q) {
		return core.Result{}, core.NewUserError("invalid full name format (expected: \"First Last\")")
	}

	parts := strings.Fields(q)
	// for now: first two tokens; later you can handle multiple last names
	first := parts[0]
	last := parts[1]

	// TODO: implement real sources lookup (file/API) in sources.go
	addr, phone, _ := "", "", "N/A (not implemented yet)"

	return core.Result{
		Kind:      core.KindFullName,
		Input:     q,
		Timestamp: time.Now(),
		FullName: core.FullNameResult{
			FirstName: first,
			LastName:  last,
			Address:   addr,
			Phone:     phone,
		},
	}, nil
}
