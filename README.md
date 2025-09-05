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

3. Run the application:
```bash
go run main.go
```

4. Open your browser and visit `http://localhost:8080`

### Command Line Options

```bash
go run main.go [options]

Options:
  -posts string     Directory containing posts (default "posts")
  -port string      HTTP listen address (default ":8080")
  -perpage int      Posts per page (default 5)
  -title string     Site title (default "Arsmp")
```

### Example Usage

```bash
# Run on custom port with different settings
go run main.go -port ":3000" -title "My Blog" -perpage 10

# Use different posts directory
go run main.go -posts "content" -port ":8080"
```

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

