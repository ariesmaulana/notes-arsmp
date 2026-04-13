# Notes Blog - Arsmp

A single-binary blog engine. Everything is embedded — just copy and run.

## Features

- **Single Binary**: Templates, static files, and posts all embedded
- **Zero Dependencies**: Download and run, that's it
- **Markdown Posts**: Write posts in markdown with frontmatter
- **Tags & Search**: Organize andfind posts easily
- **RSS Feed**: Auto-generated at `/rss`

## Quick Start

### Local Development

```bash
# Create a new post
go run main.go new "My First Post"
go run main.go new --tags "go,web" "Learning Go"

# Run locally (with hot reload)
go run main.go
```

### Build for Production

```bash
go build -ldflags="-s -w" -o blog main.go
```

That's it. Run the binary:
```bash
./blog
```

## Deployment

The binary contains everything. No config files, no static directories needed.

```bash
# Copy to server
scp blog user@server:~

# Run
./blog -port ":80" -title "My Blog"
```

## GitHub Actions Release

Push a tag to trigger automatic builds:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Downloads appear under Releases with binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Workflow

1. Write/edit posts locally in `posts/`
2. Test with `go run main.go`
3. Commit and push
4. Tag a release: `git tag v1.x.x && git push --tags`
5. GitHub Actions builds binaries
6. Download binary, copy to server, run

## Post Format

Filename: `YYYYMMDDHHMMSS-slug.md`

```markdown
title: Your Post Title
tag: tag1, tag2, tag3

Your markdown content here...
```

## Server Options

```bash
./blog -port ":3000" -perpage 10 -title "My Blog"
```