osint/
  cmd/
    osint/
      main.go                # entry point
  internal/
    cli/
      flags.go               # -fn -ip -u, --help output
      printer.go             # formatting results for terminal
    detect/
      detect.go              # detect full name / ip / username (optional)
      validators.go          # regex + parsing helpers
    services/
      fullname/
        fullname.go          # parse and query directory sources
        sources.go           # data sources (files/APIs)
      ip/
        ip.go                # IP info (ISP, city, lat/lon)
        providers.go         # provider interface + implementation(s)
      username/
        username.go          # check social networks (>=5)
        networks.go          # list of networks + rules
    output/
      writer.go              # write to result.txt/result2.txt...
      filename.go            # result naming logic
    core/
      models.go              # Result struct, common types
      errors.go              # app errors
  configs/
    osint.example.yml      # optional: API keys, endpoints
  README.md
  go.mod






to build :
go build -o osint ./cmd/osint
to run ./osint













# OSINT-Master

## Prerequisites
- Go 1.21+
- (Optional) AbuseIPDB API key for IP abuse checking
- (Optional) SecurityTrails API key for enhanced subdomain enumeration

## Installation
```bash
git clone &lt;repo&gt;
cd osint-master
go mod tidy
go build -o osintmaster cmd/osintmaster/main.go