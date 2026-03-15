## OSINT-Master

<center>
<img src="./resources/osint-meme.png?raw=true" style="width: 673px !important; height: 439px !important;"/>
</center>

### Introduction

Open-source intelligence (OSINT) is a key component of cybersecurity, providing valuable insights into potential vulnerabilities and security risks. This project involves creating a tool that performs comprehensive passive reconnaissance using publicly available data.

### Objective

The goal is to build a multi-functional tool using a programming language of your choice. The tool is capable of retrieving detailed information based on user inputs such as `IP addresses`, `usernames`, and `domains`. This project will enhance your skills in data analysis, ethical considerations, and the use of various cybersecurity tools and APIs.

By completing this project, you will:

- Develop an understanding of OSINT techniques and their applications.
- Gain practical experience in programming, API integration, and data handling.
- Learn to identify and mitigate security risks, including subdomain takeovers.
- Understand the ethical and legal implications of cybersecurity practices.

### Resources

Some useful resources:

- [Open-Source Intelligence](https://en.wikipedia.org/wiki/Open-source_intelligence)
- [Doxing](https://en.wikipedia.org/wiki/Doxing)
- [OSINT Tools on GitHub](https://github.com/topics/osint-tools)
- [OSINT Framework](https://osintframework.com/) - Comprehensive collection of OSINT tools and resources
- [Awesome OSINT](https://github.com/jivoi/awesome-osint) - Curated list of OSINT resources and tools
- [OSINT Techniques](https://www.osinttechniques.com/) - Resources and techniques for OSINT investigations
- [IntelTechniques](https://inteltechniques.com/tools/index.html) - Collection of OSINT search tools

Before asking for help, ask yourself if you have really thought about all the possibilities.

### Role Play

To enhance the learning experience and assess your knowledge, a role-play question session will be included as part of this project. This section will involve answering a series of questions in a simulated real-world scenario where you assume the role of a Cyber Security Expert explaining how to protect information from OSINT techniques to a team or stakeholder.

The goal of the role-play question session is to:

- Assess your understanding of OSINT risks and mitigation strategies.
- Test your ability to communicate effectively and explain security measures related to this project.
- Challenge you to think critically about the importance of information security and consider alternative approaches.
- Explain what subdomain takeovers are and how to protect against them.

Prepare for a role-play question session during the audit.

### Development Environment

**Virtual Machine Recommendation:**

For security and isolation purposes, it is **strongly recommended** to develop and test this tool in a virtual machine (VM):

- **Why use a VM:**
  - Isolates OSINT activities from your personal system
  - Protects API keys and sensitive credentials
  - Provides a clean testing environment
  - Allows safe interaction with potentially risky domains/IPs
  - Facilitates network traffic monitoring and analysis

- **Recommended Setup:**
  - Linux-based VM (Ubuntu 20.04+ or Kali Linux)
  - Minimum 2GB RAM, 20GB disk space
  - Network adapter configured for NAT or Bridged mode
  - Snapshot capability for easy rollback

- **VM Software Options:**
  - VirtualBox (free, cross-platform)
  - VMware Workstation Player (free for personal use)
  - QEMU/KVM (Linux hosts)

### Project Requirements

#### Input Handling

The tool should accept the following inputs: `IP Address`, `Username`, and `Domain`.

#### Information Retrieval

- **IP Address:**
  Retrieve geolocation data, ISP details, and check for any historical data associated with the IP (e.g., from abuse databases).

- **Username:**
  Check for the presence of the username on at least five known social networks and public repositories.
  Retrieve public profile information, such as profile bio, activity status, and follower count.

- **Domain and Subdomain Enumeration:**
  Enumerate subdomains and gather information including IP addresses, SSL certificate details, and potential vulnerabilities.
  Identify potential subdomain takeover risks by analyzing DNS records and associated resources.

> You are responsible for choosing the data sources and APIs you want to use. Be aware of each API's `Terms of Use` and `Cost` before use. Using APIs and external data sources is expected, but simply wrapping an existing OSINT CLI tool (such as `theHarvester`, `Sherlock`, or `subfinder`) without implementing your own logic is not acceptable.

#### Output Management

Store the results in a well-organized file format.

### Usage Examples

#### Command Line Interface

```sh
$> osintmaster --help

Welcome to osintmaster multi-function Tool

OPTIONS:
    -i  "IP Address"       Search information by IP address
    -u  "Username"         Search information by username
    -d  "Domain"           Enumerate subdomains and check for takeover risks
    -o  "FileName"         File name to save output
    --help                 Display this help message
```

#### Example Outputs

**IP Address:**

```sh
$> osintmaster -i 8.8.8.8 -o result1.txt
ISP: Google LLC
City: Mountain View
Country: COUNTRY
ASN: 15169
Known Issues: No reported abuse
Data saved in result1.txt
```

**Username:**

```sh
$> osintmaster -u "@username" -o result2.txt
Facebook: Found
Twitter: Found
LinkedIn: Found
Instagram: Not Found
GitHub: Found
Recent Activity: Active on GitHub, last post 1 day ago
Data saved in result2.txt
```

**Domain and Subdomain Enumeration:**

```sh
$> osintmaster -d "example.com" -o result3.txt
Main Domain: example.com

Subdomains found: 3
  - www.example.com (IP: 123.123.123.123)
    SSL Certificate: Valid until 2030-03-01
  - mail.example.com (IP: 123.123.123.123)
    SSL Certificate: Valid until 2030-03-01
  - test.example.com (IP: 123.123.123.123)
    SSL Certificate: Not found

Potential Subdomain Takeover Risks:
  - Subdomain: test.example.com
    CNAME record points to a non-existent AWS S3 bucket
    Recommended Action: Remove or update the DNS record to prevent potential misuse

Data saved in result3.txt
```

### Bonus

If you complete the mandatory part successfully and still have free time, you can implement anything that you feel deserves to be a bonus. For example:

- **Full Name Search:** Add a `-n` option to search information by full name, including phone numbers, addresses, and social media profiles.
- **User Interface:** Develop a graphical user interface (GUI) for better user accessibility.
- **PDF Generation:** Add a feature to generate your OSINT results as PDF files.

Challenge yourself!

### Documentation

Create a `README.md` file that provides comprehensive documentation for your tool (prerequisites, setup, configuration, usage, etc.). This file must be submitted as part of the solution for the project.

Add clear guidelines and warnings about the ethical and legal use of the tool to your documentation.

### Ethical and Legal Considerations

- **Get Permission:** Always obtain explicit permission before gathering information.
- **Respect Privacy:** Collect only necessary data and store it securely.
- **Follow Laws:** Adhere to relevant laws such as GDPR and CFAA.
- **Report Responsibly:** Privately notify affected parties of any vulnerabilities.
- **Educational Use Only:** Use this tool and techniques solely for learning and improving security.

> **Disclaimer:** This project is for educational purposes only. Ensure all activities comply with legal and ethical standards. The institution is not responsible for misuse of the techniques and tools demonstrated.

### Repository Structure

Your repository should be organized as follows:

```
osint-master/
├── src/
│   ├── ip_lookup.py (or your language equivalent)
│   ├── username_lookup.py
│   ├── domain_enum.py
│   └── main.py (or osintmaster entry point)
├── tests/
│   └── test_*.py (optional test files)
├── output/
│   └── (directory for storing results)
├── resources/
│   └── osint-meme.png
├── README.md
├── requirements.txt (or equivalent for your language)
└── .gitignore
```

**Note:** The exact structure may vary depending on your programming language and implementation approach, but ensure that:

- Source code is well-organized in appropriate directories
- All necessary files for running the tool are included
- The project structure is clearly documented in the README.md

### Submission Requirements

For the audit, you must submit a complete repository containing:

1. **README.md**: Comprehensive documentation including:
   - Project overview and objectives
   - Prerequisites and dependencies
   - Installation and setup instructions
   - Usage examples for each feature
   - Command-line options and parameters
   - Output format explanations
   - API configuration (if applicable)
   - Ethical and legal guidelines
   - Troubleshooting tips
   - Known limitations

2. **Source Code**: Well-organized and commented code for all features:
   - IP Address lookup
   - Username search
   - Domain and subdomain enumeration

3. **Configuration Files**: Any necessary configuration files (e.g., `requirements.txt`, `package.json`, `Makefile`, `Cargo.toml`)

4. **Output Directory**: Pre-created directory structure for storing results

**Important Notes:**

- All code must be your own work (implementing your own logic)
- Simply wrapping existing OSINT CLI tools is not acceptable
- Code should be well-commented and follow best practices for your chosen language
- Include error handling and input validation
- Document API usage and any rate limits

### Submission and Audit

Upon completing this project, you should be prepared to:

- Demonstrate all features working correctly
- Explain your implementation decisions and code architecture
- Describe your API integration approach
- Participate in a role-play session as a Cyber Security Expert
- Discuss ethical and legal considerations of OSINT
- Show understanding of the underlying concepts and techniques
- Explain subdomain takeover risks and mitigation strategies

The audit will verify that:

- All required files are present and properly organized
- Tool is implemented with custom logic, not just wrapping existing tools
- Documentation is comprehensive and accurate
- You can effectively communicate your knowledge and decisions
- You understand ethical considerations and privacy concerns
