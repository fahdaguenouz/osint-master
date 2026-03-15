#### General

##### Development Environment Verification

###### Did the learner develop and test the tool in an isolated environment (VM recommended)?

###### Can the learner explain why isolation is important for OSINT tool development?

**Note for Auditors:** While not mandatory, verify that the learner understands the security implications of running OSINT tools and has taken appropriate precautions. The audit itself can be conducted on the auditor's VM or secure testing environment.

##### Check the Repo Content

Files that must be inside the repository:

- Detailed documentation in the `README.md` file.
- Source code for the OSINT-Master tool.
- Any required configuration files and scripts for running the tool.

###### Are all the required files present?

##### Repository Structure Verification

###### Is the repository well-organized with a clear directory structure?

###### Are source code files properly organized in appropriate directories (e.g., `src/`, `lib/`, or equivalent)?

###### Is there a dedicated directory or location for output files?

###### Are dependencies clearly documented (e.g., `requirements.txt`, `package.json`, `Gemfile`, `Cargo.toml`, or equivalent)?

##### Play the Role of a Stakeholder

Organize a simulated scenario where the learner takes on the role of a Cyber Security Expert and explains their solution and knowledge to a team or stakeholder. Evaluate their grasp of the concepts and technologies used in the project, their communication efficacy, and their critical thinking about their solution and the knowledge behind this project.

Suggested role-play questions include:

- What is OSINT, and why is it significant in cybersecurity?
- What types of information can be gathered using OSINT techniques?
- Explain what subdomain takeovers are and how to protect against them.
- How does the OSINT-Master tool help identify sensitive information?
- What challenges did you face while developing the OSINT-Master tool, and how did you address them?
- How can we protect our critical information from OSINT techniques?
- How can this tool help in a defensive approach?

###### Was the learner able to answer all the questions?

###### Did the learner demonstrate a thorough understanding of the concepts and technologies used in the project?

###### Was the learner able to communicate effectively, justify their decisions, and explain the knowledge behind this project?

###### Was the learner able to evaluate the value of this project in real-life scenarios?

###### Did the learner demonstrate an understanding of ethical and legal considerations related to OSINT?

##### Check the Learner Documentation in the `README.md` File

###### Does the `README.md` file contain all the necessary information about the tool (prerequisites, setup, configuration, usage, etc.)?

###### Does the `README.md` include clear installation and setup instructions?

###### Does the `README.md` provide usage examples for each feature (IP lookup, username search, domain enumeration)?

###### Does the `README.md` explain the command-line options and parameters?

###### Does the `README.md` describe the output format and where results are stored?

###### Does the `README.md` document API configuration and usage (if applicable)?

###### Does the `README.md` file contain clear guidelines and warnings about the ethical and legal use of the tool?

###### Does the `README.md` include troubleshooting information or common issues?

###### Does the `README.md` document any known limitations (e.g., API rate limits, data source restrictions)?

##### Review the Tool's Design and Implementation

1. **Help Command:**

```sh
$> osintmaster --help
```

###### Does the output include an explanation of how to use the tool?

2. **IP Address Option:**

```sh
$> osintmaster -i "IP Address" -o filename
```

###### Does the output include geolocation data, ISP details, and historical data?

###### Is the output stored in the file specified in the output parameter?

###### Does the output file contain properly formatted results?

3. **Username Option:**

```sh
$> osintmaster -u "Username" -o filename
```

###### Does the output check the presence of the username on at least five social networks and public repositories?

###### Does the output retrieve public profile information (bio, activity status, follower count)?

###### Is the output stored in the file specified in the output parameter?

###### Does the output file contain properly formatted results?

4. **Domain Option:**

```sh
$> osintmaster -d "Domain" -o filename
```

###### Does the output enumerate subdomains, gather relevant information, and identify potential subdomain takeover risks?

###### Does the output include IP addresses and SSL certificate details for subdomains?

###### Is the output stored in the file specified in the output parameter?

###### Does the output file contain properly formatted results?

##### Code Quality and Implementation

###### Is the code well-organized and properly structured?

###### Does the code include meaningful comments explaining complex logic?

###### Does the code follow best practices for the chosen programming language?

###### Is there proper error handling for common failure scenarios (e.g., API errors, network timeouts, invalid input)?

###### Does the code validate user input appropriately?

###### Does the code handle API rate limits gracefully?

##### Ethical and Privacy Considerations

###### Did the learner demonstrate responsible use of OSINT techniques?

###### Can the learner explain the privacy implications of gathering publicly available information?

###### Does the learner understand the legal considerations of OSINT activities?

###### Can the learner describe how to use OSINT defensively to protect information?

##### Ensure That the Learner Submission Meets the Project Requirements

1. **Functionality:** Does the tool retrieve detailed information based on the given inputs (IP Address, Username, and Domain)?

2. **Data Accuracy:** Is the retrieved information accurate and relevant?

3. **Ethical Considerations:** Are there clear guidelines and warnings about the ethical and legal use of the tool?

4. **Usability:** Is the tool user-friendly and well-documented?

5. **Custom Implementation:** Did the learner implement their own logic rather than simply wrapping existing tools?

> Note: Using APIs and external data sources is expected. However, simply wrapping an existing OSINT CLI tool (such as `theHarvester`, `Sherlock`, or `subfinder`) without implementing custom logic is not acceptable. If an API is unavailable during the audit, the learner should demonstrate that their implementation works correctly and explain their approach.

###### Did the learner implement the logic and integration themselves rather than simply wrapping an existing OSINT CLI tool?

###### Did the tool design and implementation align with all the project requirements above?

###### Was the learner able to implement a functional and reliable tool that meets the project requirements?

###### Can the learner explain their implementation approach and architecture decisions?

###### Can the learner demonstrate their tool working correctly?

##### Submission Completeness

###### Does the submission include all required source code files?

###### Does the submission include all necessary configuration and dependency files?

###### Are all files properly committed to the repository (no missing dependencies)?

###### Can the tool be run successfully after following the setup instructions in the README?

###### Does the submission include proper documentation about data sources and APIs used?

#### Bonus

###### + Did the learner implement a Full Name search feature (`-n` option) that retrieves phone numbers, addresses, and social media profiles?

###### + Did the learner implement additional valuable features (GUI, PDF generation, etc.)?

###### + Did the learner include comprehensive testing or test files?

###### + Does the project demonstrate exceptional code quality and organization?

###### + Did the learner implement additional security analysis features?

###### + Is this project an outstanding project that exceeds the basic requirements?
