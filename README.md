
## `README.md`

```markdown
# OSINT-Master

A comprehensive open-source intelligence (OSINT) tool for passive reconnaissance, built in Go.

![OSINT Meme](resources/osint-meme.png)

## Overview

OSINT-Master is a multi-functional command-line tool that performs passive reconnaissance using publicly available data sources. It retrieves detailed information based on user inputs including IP addresses, usernames, and domains.

**⚠️ Educational Use Only**: This tool is designed for educational purposes and authorized security testing only.

## Features

- **IP Address Lookup**: Geolocation, ISP details, ASN, and abuse reputation checking
- **Username Search**: Profile discovery across 5+ social networks (GitHub, Twitter/X, Instagram, TikTok, Facebook) with bio, follower counts, and recent activity
- **Domain Enumeration**: Subdomain discovery, SSL certificate validation, and subdomain takeover risk detection
- **Full Name Search** (Bonus): Basic full name parsing (extensible for future enhancement)

## Prerequisites

- Go 1.21 or higher
- Internet connection
- (Optional) AbuseIPDB API key for enhanced IP reputation checking

## Installation

```bash
# Clone the repository
git clone https://github.com/fahdaguenouz/osint
cd osint

# Install dependencies
go mod tidy

# Build the tool
go build -o osintmaster cmd/osintmaster/main.go

# Or run directly without building
go run cmd/osintmaster/main.go --help
```

## Configuration

### Optional: AbuseIPDB API Key

For enhanced IP abuse checking, set your API key:

```bash
export ABUSEIPDB_API_KEY="your_api_key_here"
```

Free tier: 1,000 checks per day. The tool works without this, but won't show abuse confidence scores.

## Usage

### Help

```bash
./osintmaster --help
```

Output:
```
Welcome to osintmaster multi-function Tool

OPTIONS:
    -i  "IP Address"       Search information by IP address
    -u  "Username"         Search information by username
    -d  "Domain"           Enumerate subdomains and check for takeover risks
    -n  "Full Name"        Search information by full name (bonus)
    -o  "FileName"         File name to save output
    --help                 Display this help message
```

### IP Address Lookup

```bash
./osintmaster -i 8.8.8.8 -o result.txt
```

**Example Output:**
```
ISP: Google LLC
City: Mountain View
Country: United States
ASN: AS15169 Google LLC
Lat/Lon: 37.3860 / -122.0838
Known Issues: No reported abuse

Data saved in result.txt
```

**Data Sources:**
- ip-api.com (free, no API key): Geolocation, ISP, ASN
- AbuseIPDB (optional): Abuse confidence score and report count

### Username Search

```bash
./osintmaster -u torvalds -o result.txt
```

**Example Output:**
```
Facebook: Not Found
Twitter: Found (2.5M followers)
  Bio: Creator of Linux...
Instagram: Not Found
Tiktok: Not Found
Github: Found (100k+ followers)
  Bio: Linux creator
  Recent Activity:
    - Repository: linux (2024-03-15)
    - Repository: git (2024-03-10)

Recent Activity: Active on: twitter, github
Last Post: Repository: linux on GitHub (2024-03-15)

Data saved in result.txt
```

**Platforms Checked:**
1. GitHub - Public repositories, bio, followers
2. Twitter/X - Bio, followers, recent tweets
3. Instagram - Bio, followers, recent posts
4. TikTok - Bio, followers, recent videos
5. Facebook - Basic profile info (limited due to privacy restrictions)

**Note:** Social media platforms frequently change their anti-scraping measures. Some platforms may require login or show rate limits.

### Domain Enumeration

```bash
./osintmaster -d example.com -o result.txt
```

**Example Output:**
```
Main Domain: example.com

Subdomains found: 3
  - example.com (IP: 93.184.216.34)
    SSL Certificate: Valid until 2025-12-01
  - www.example.com (IP: 93.184.216.34)
    SSL Certificate: Valid until 2025-12-01
  - mail.example.com (IP: 93.184.216.35)
    SSL Certificate: Valid until 2025-12-01

Potential Subdomain Takeover Risks: None detected

Data saved in result.txt
```

**Features:**
- Subdomain enumeration via Certificate Transparency logs (crt.sh)
- DNS resolution for each subdomain
- SSL certificate validation and expiry dates
- Subdomain takeover detection (dangling CNAMEs to cloud services)

**Detected Takeover Services:**
- AWS S3, CloudFront
- GitHub Pages
- Heroku
- Azure (Websites, Blob Storage)
- Vercel, Netlify, Cloudflare Pages
- And more...

### Full Name Search (Bonus)

```bash
./osintmaster -n "John Doe" -o result.txt
```

Currently returns parsed first/last name. Extensible for future data source integration.

## Output Format

Results are saved in plain text format with:
- Timestamp
- Query type and input
- Structured results per feature
- Data sources used
- Warnings (if any)

Files are saved as `result.txt`, `result2.txt`, etc., or custom filename via `-o` flag.

## Architecture

```
osint/
├── cmd/osintmaster/
│   └── main.go              # Entry point
├── internal/
│   ├── cli/
│   │   ├── flags.go         # Command-line parsing
│   │   └── printer.go       # Terminal output
│   ├── core/
│   │   └── models.go        # Data structures
│   ├── services/
│   │   ├── domain/          # Domain enumeration
│   │   ├── ip/              # IP geolocation & abuse
│   │   ├── username/        # Social media checks
│   │   └── fullname/        # Full name parsing
│   └── output/
│       └── writer.go        # File output
├── resources/
│   └── osint-meme.png
├── go.mod
└── README.md
```

## API Usage & Rate Limits

| Service | Type | Limits | Auth Required |
|---------|------|--------|---------------|
| ip-api.com | IP Geolocation | 45 requests/minute | No |
| AbuseIPDB | IP Reputation | 1,000/day | Optional (free tier) |
| crt.sh | Subdomain Enum | No limit | No |
| GitHub API | Profile/Repos | 60/hour (unauthenticated) | No |
| Social Media | Profile Check | Varies | No |

**Rate Limit Handling:**
- The tool implements timeouts and backoff strategies
- Warnings are displayed if rate limits are hit
- Results are cached where possible to minimize requests

## Ethical and Legal Guidelines

### Responsible Use

1. **Get Permission**: Always obtain explicit permission before gathering information about individuals or organizations.

2. **Respect Privacy**: Collect only necessary data and store it securely. Do not share or publish personal information.

3. **Follow Laws**: Adhere to relevant laws such as:
   - GDPR (EU)
   - CFAA (US)
   - Local privacy and data protection regulations

4. **Report Responsibly**: If you discover vulnerabilities (e.g., subdomain takeovers), privately notify the affected parties.

5. **Educational Use**: This tool is for learning and authorized security testing only.

### What NOT to Do

- Do not use this tool for stalking, harassment, or doxxing
- Do not attempt to claim resources during subdomain takeover checks (passive detection only)
- Do not circumvent rate limits or terms of service
- Do not store or distribute personal data without consent

## Troubleshooting

### Common Issues

**"No result" or timeouts**
- Check internet connection
- Some services may be rate-limited; wait a few minutes
- Try with a VPN if your IP is blocked

**TikTok/Instagram always returns "Not Found"**
- These platforms aggressively block automated requests
- Try different usernames or wait between requests
- Consider using residential proxies (not included in this tool)

**AbuseIPDB shows "No API key"**
- Set `export ABUSEIPDB_API_KEY="your_key"`
- Or ignore - the tool works without it, just without abuse scores

**SSL certificate errors**
- Some sites use self-signed certificates; these are noted as "Not found"

## Known Limitations

1. **Social Media Blocking**: Platforms like Instagram, TikTok, and Facebook frequently change their anti-bot measures and may block requests or require login.

2. **Rate Limiting**: Unauthenticated GitHub API requests are limited to 60/hour.

3. **IP Geolocation Accuracy**: IP geolocation is approximate (city-level accuracy varies by ISP).

4. **Subdomain Takeover Verification**: Detection is passive (DNS-based). Actual vulnerability verification would require attempting to claim resources, which this tool does NOT do.

5. **Full Name Search**: Currently basic; comprehensive people search requires paid APIs or databases.

## Development

### Testing

```bash
# Run all tests
go test ./...

# Test specific module
go test ./internal/services/ip/
```

### Adding New Features

The modular architecture makes it easy to add:
- New social networks in `internal/services/username/networks.go`
- New IP providers in `internal/services/ip/providers.go`
- New subdomain sources in `internal/services/domain/`

## License

This project is for educational purposes. Use responsibly and ethically.

## Author

[Fahd Aguenouz]

---

**Disclaimer**: The authors are not responsible for misuse of this tool. Always ensure you have proper authorization before conducting OSINT investigations.
```

---

After creating the README, the remaining steps are:

1. **Create `go.mod`** if not exists - ensure it has proper module name
2. **Add `resources/osint-meme.png`** - copy any relevant OSINT meme image
3. **Create `.gitignore`**:
```gitignore
*.txt
result*
!resources/*.txt
osintmaster
*.exe
.env
```

4. **Final testing** of all features:
```bash
./osintmaster --help
./osintmaster -i 8.8.8.8 -o ip_test.txt
./osintmaster -u torvalds -o user_test.txt
./osintmaster -d github.com -o domain_test.txt
```


go get github.com/playwright-community/playwright-go
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium