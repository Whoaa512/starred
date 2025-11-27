# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go CLI tool that generates an "Awesome List" README from a GitHub user's starred repositories. It fetches starred repos via GitHub's GraphQL API, groups them by language (or topic), and either outputs the markdown or commits it directly to a specified repository.

## Commands

```bash
# Run the tool locally
go run main.go --username <user> --token <github_token> --sort

# Run with repository output (commits to GitHub)
go run main.go --username <user> --repository <repo> --sort --token <token> --message 'commit message'

# Build
go build -o starred main.go
```

## Key CLI Flags

- `--username/-u`: GitHub username (required)
- `--token/-t`: GitHub personal access token (required)
- `--repository/-r`: Target repo to commit README to (if omitted, prints to stdout)
- `--sort/-s`: Sort categories alphabetically
- `--topic`: Group by topic instead of language
- `--topic_limit`: Minimum stargazer count for topic inclusion (default: 500)
- `--private/-p`: Include private repos

## Architecture

Single-file Go application (`main.go`) with:

- `GitHubClient`: Wraps GraphQL (for starred repo queries) and REST (for file operations) clients
- `GetAllStarredRepositories`: Paginated fetch with rate-limit handling and 3s delays between pages
- `generateREADME`: Builds markdown with table of contents, grouped by language/topic
- `CreateOrUpdateFile`: Creates repo if needed, then creates/updates the README file

The tool is designed to run via GitHub Actions on a cron schedule (daily at 6:30 UTC).
