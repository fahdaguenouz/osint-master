package main

import (
	"fmt"
	"os"
	"path/filepath"

	"osint/src/cli"
	"osint/src/core"
	"osint/src/output"
	"osint/src/services/domain"
	"osint/src/services/ip"
	"osint/src/services/username"
)

func main() {
	opts, showHelp, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		cli.PrintHelp(os.Stderr)
		os.Exit(2)
	}
	if showHelp {
		cli.PrintHelp(os.Stdout)
		return
	}

	var (
		res    core.Result
		runErr error
	)

	switch opts.Mode {
	case cli.ModeIP:
		res, runErr = ip.Run(opts.Query)
	case cli.ModeUsername:
		res, runErr = username.Run(opts.Query)
	case cli.ModeDomain:
		res, runErr = domain.Run(opts.Query)
	default:
		fmt.Fprintln(os.Stderr, "Error: no mode selected (-n, -i, -u, -d)")
		cli.PrintHelp(os.Stderr)
		os.Exit(2)
	}


	if runErr != nil {
        fmt.Fprintf(os.Stderr, "Execution failed: %v\n", runErr)
        os.Exit(1) // Exit immediately before trying to print or save empty results
    }
	// Always print results 
	cli.PrintResult(os.Stdout, res)

	// Determine output filename
	filename := opts.Output

	// If no -o specified, auto-generate in results/ folder
	if filename == "" {
		filename, err = output.NextResultFilename("results")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	} else {
		// If -o provided with just a filename (no path), put it in results/
		if filepath.Dir(filename) == "." || filepath.Dir(filename) == "" {
			filename = filepath.Join("results", filename)
		}
		// Ensure the directory exists for the specified path
		if dir := filepath.Dir(filename); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintln(os.Stderr, "Error creating directory:", err)
				os.Exit(1)
			}
		}
	}

	if err := output.WriteResult(filename, res); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Data saved in %s\n", filename)
}
