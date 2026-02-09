# Notes Blog - Arsmp

A lightweight, fast blog engine written in Go that serves markdown posts with automatic file watching and hot reloading.

## Features

- **Markdown Posts**: Write posts in markdown format with frontmatter metadata
- **Search**: Full-text search across post titles and tags
- **Tags**: Organize posts with tags and browse by tag
- **RSS Feed**: Automatically generated RSS feed for recent posts at `/rss`
- **Pagination**: Configurable posts per page
- **Hot Reload**: Automatic post reloading when files change
- **Clean UI**: Minimal, responsive design

## Quick Start

### Prerequisites

- Go 1.23.4 or later
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd notes-arsmp
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o blog main.go
```

## Usage

### Creating a New Post

Use the `new` command to quickly create a new blog post:

```bash
# Create a post with title only
./blog new "My Awesome Post"

# Create a post with tags
./blog new --tags "python,django,web" "Django Tips and Tricks"

# Using go run
go run main.go new "Another Post Title"
```

This will:
- Generate a file with the format: `YYYYMMDDHHMMSS-slug.md`
- Create a slug from the title (lowercase, hyphens for spaces)
- Pre-fill the frontmatter with title and tag fields
- Save the file in the `posts` directory

Example output:
```
Created: posts/20260126143000-my-awesome-post.md
```

The generated post file will have this structure:
```markdown
title: My Awesome Post
tag: python,django,web

[Your content goes here]
```

### Running the Server

Start the web server to view your blog:

```bash
# Run with defaults (port :8080)
./blog serve

# Or simply (serve is the default command)
./blog

# Run with custom settings
./blog serve -port ":3000" -title "My Blog" -perpage 10

# Using go run
go run main.go serve
```
Open your browser and visit `http://localhost:8080`

### Command Reference

```bash
# Post creation command
./blog new [--tags tag1,tag2] <title>
./blog new --posts "content" "My Post"    # Custom posts directory

# Server command  
./blog serve [options]
./blog serve -port ":3000"                # Custom port
./blog serve -perpage 10                  # Posts per page
./blog serve -title "My Blog"             # Site title
./blog serve -posts "content"             # Custom posts directory

# Default behavior (backward compatible)
./blog                                     # Runs serve with defaults
./blog -port ":3000"                       # Runs serve with custom port
```

### Post Format

Posts are markdown files with frontmatter metadata. The filename format is:
- `YYYYMMDD-slug.md` (date only)
- `YYYYMMDDHHMMSS-slug.md` (date and time)

Frontmatter format:
```markdown
title: Your Post Title
tag: tag1, tag2, tag3

Your markdown content starts here...
```

## Development Workflow

1. Create a new post:
   ```bash
   ./blog new --tags "tutorial,go" "Getting Started with Go"
   ```

2. Edit the generated file in your favorite editor

3. Start the development server:
   ```bash
   go run main.go
   ```

4. The server will automatically reload when you modify posts

## Command Line Options (Legacy)

For backward compatibility, you can still use the original command format:

```bash
go run main.go [options]

Options:
  -posts string     Directory containing posts (default "posts")
  -port string      HTTP listen address (default ":8080")
  -perpage int      Posts per page (default 5)
  -title string     Site title (default "Arsmp")
```

Note: The new `serve` subcommand is recommended for clarity.

## RSS Feed

The blog automatically generates an RSS feed available at `/rss`. The feed includes:

- Recent 20 posts (configurable in the code)
- Post title, description, and publication date
- Post tags as categories
- Proper RSS 2.0 format with Atom extensions

You can subscribe to the RSS feed using any RSS reader by navigating to `http://localhost:8080/rss` (or your domain + `/rss` in production).

## Building and Deployment

### Building for Production

For production deployment, you should build a binary instead of running with `go run`:

```bash
# Build the binary
go build -o blog main.go

# Run the binary directly
./blog -port ":8080" -title "My Production Blog"
```

#### Cross-compilation for Different Platforms

```bash
# Build for Linux (common for servers)
GOOS=linux GOARCH=amd64 go build -o blog-linux main.go

# Build for different architectures
GOOS=linux GOARCH=arm64 go build -o blog-arm64 main.go
```

#### Optimized Production Build

```bash
# Build with optimizations (smaller binary, no debug info)
go build -ldflags="-s -w" -o blog main.go
```

