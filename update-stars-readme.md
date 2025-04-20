# Starred Go

A Go port of the [starred](https://github.com/maguowei/starred) Python module.

This tool creates your own Awesome List of GitHub stars!

## Installation

```bash
go install github.com/maguowei/starred@latest
```

## Usage

```bash
starred --username <your-github-username> --token <your-github-token> --sort
```

### Options

- `--username`, `-u`: GitHub username (required)
- `--token`, `-t`: GitHub token (required)
- `--sort`, `-s`: Sort by category name alphabetically
- `--topic`: Category by topic, default is category by language
- `--topic_limit`: Topic stargazer count gt number, set bigger to reduce topics number (default: 500)
- `--repository`, `-r`: Repository name
- `--filename`, `-f`: File name (default: "README.md")
- `--message`, `-m`: Commit message (default: "update awesome-stars, created by starred")
- `--private`, `-p`: Include private repos

## Example

```bash
# Export your awesome lists
starred --username <your-github-username> --token <your-github-token> --sort > README.md

# Or using environment variables
export GITHUB_TOKEN=<your-github-token>
export USER=<your-github-username>
starred --sort > README.md
```

## License

MIT