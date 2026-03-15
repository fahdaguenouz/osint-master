package cli

import (
	"fmt"
	"io"
	"strings"

	"osint/internal/core"
)

func PrintResult(w io.Writer, r core.Result) {
	if r.Error != "" {
		fmt.Fprintf(w, "Error: %s\n", r.Error)
		return
	}

	for _, warn := range r.Warnings {
		fmt.Fprintf(w, "Warning: %s\n", warn)
	}

	switch r.Kind {
	case core.KindFullName:
		fmt.Fprintf(w, "First name: %s\n", r.FullName.FirstName)
		fmt.Fprintf(w, "Last name: %s\n", r.FullName.LastName)
		if r.FullName.Address != "" {
			fmt.Fprintf(w, "Address: %s\n", r.FullName.Address)
		}
		if r.FullName.Phone != "" {
			fmt.Fprintf(w, "Number: %s\n", r.FullName.Phone)
		}

	case core.KindIP:
		if r.IP.ISP != "" {
			fmt.Fprintf(w, "ISP: %s\n", r.IP.ISP)
		}
		if r.IP.City != "" {
			fmt.Fprintf(w, "City: %s\n", r.IP.City)
		}
		if r.IP.Lat != 0 || r.IP.Lon != 0 {
			fmt.Fprintf(w, "City Lat/Lon:\t(%.4f) / (%.4f)\n", r.IP.Lat, r.IP.Lon)
		}

	case core.KindUsername:
		for _, n := range r.Username.Networks {
			val := "no"
			if n.Found {
				val = "yes"
			}

			name := n.Name
			if len(name) > 0 {
				name = strings.ToUpper(name[:1]) + name[1:]
			}

			fmt.Fprintf(w, "%s : %s\n", name, val)
		}

	default:
		fmt.Fprintln(w, "No result.")
	}
}
