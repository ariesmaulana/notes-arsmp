package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
)

const (
	defaultPostsDir  = "posts"
	defaultPerPage   = 5
	defaultPort      = ":8080"
	dateLayoutDay    = "20060102"
	dateLayoutSecond = "20060102150405"
)

// Data structure

type PostMeta struct {
	Title    string
	Slug     string
	Date     time.Time
	Tags     []string
	Filename string
}

type Post struct {
	Meta PostMeta
	HTML template.HTML
}
type App struct {
	Posts    []PostMeta
	BySlug   map[string]int
	TagIndex map[string][]int

	PerPage   int
	Templates *template.Template
	SiteTitle string
	PostsDir  string
	Markdown  goldmark.Markdown

	mu sync.RWMutex
}

func (a *App) reloadPosts() error {
	posts, err := loadPosts(a.PostsDir)
	if err != nil {
		return err
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	bySlug := make(map[string]int, len(posts))
	tagIndex := make(map[string][]int)
	for i, p := range posts {
		bySlug[p.Slug] = i
		for _, tag := range p.Tags {
			tagIndex[tag] = append(tagIndex[tag], i)
		}
	}

	a.mu.Lock()
	a.Posts = posts
	a.BySlug = bySlug
	a.TagIndex = tagIndex
	a.mu.Unlock()

	log.Printf("[reload] posts reloaded: %d posts", len(posts))
	return nil
}

func (a *App) watchPosts() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("[watcher] error init: %v", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(a.PostsDir); err != nil {
		log.Printf("[watcher] cannot watch dir: %v", err)
		return
	}
	log.Printf("[watcher] watching directory: %s", a.PostsDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			switch {
			case event.Op&fsnotify.Create != 0:
				log.Printf("[watcher] new file: %s", event.Name)
			case event.Op&fsnotify.Write != 0:
				log.Printf("[watcher] modified: %s", event.Name)
			case event.Op&fsnotify.Remove != 0:
				log.Printf("[watcher] removed: %s", event.Name)
			case event.Op&fsnotify.Rename != 0:
				log.Printf("[watcher] renamed: %s", event.Name)
			}
			if err := a.reloadPosts(); err != nil {
				log.Printf("[watcher] reload error: %v", err)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[watcher] error: %v", err)
		}
	}
}

var filenameRe = regexp.MustCompile(`^(\d{8}|\d{14})-(.+?)\.md$`)

func main() {
	if len(os.Args) < 2 {
		runServe(os.Args)
		return
	}

	switch os.Args[1] {
	case "new":
		runNew(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	default:
		runServe(os.Args[1:])
	}
}

func runNew(args []string) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	postsDir := fs.String("posts", defaultPostsDir, "Directory containing posts")
	tags := fs.String("tags", "", "Comma-separated tags")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("[new] error parsing flags: %v", err)
	}

	if fs.NArg() < 1 {
		log.Fatal("[new] error: post title is required\nUsage: new [--tags tag1,tag2] <title>")
	}

	title := strings.Join(fs.Args(), " ")
	slug := generateSlug(title)
	timestamp := time.Now().Format(dateLayoutSecond)
	filename := fmt.Sprintf("%s-%s.md", timestamp, slug)
	filePath := filepath.Join(*postsDir, filename)

	tagValue := ""
	if *tags != "" {
		tagValue = *tags
	}
	content := fmt.Sprintf("title: %s\ntag: %s\n\n", title, tagValue)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Fatalf("[new] error creating file: %v", err)
	}

	fmt.Printf("Created: %s\n", filePath)
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	slug = strings.Trim(slug, "-")

	return slug
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	postDir := fs.String("posts", defaultPostsDir, "Directory containing post")
	port := fs.String("port", defaultPort, "HTTP listen address")
	perPage := fs.Int("perpage", defaultPerPage, "Post per page")
	siteTitle := fs.String("title", "Arsmp", "Site Title")
	fs.Parse(args)

	app, err := NewApp(*postDir, *perPage, *siteTitle)
	if err != nil {
		log.Fatalf("[init] error: %v", err)
	}
	r := chi.NewRouter()
	r.Get("/", app.handleIndex)
	r.Get("/page/{n}", app.handleIndexPage)
	r.Get("/post/{slug}", app.handlePost)
	r.Get("/tag/{tag}", app.handleTag)
	r.Get("/search", app.handleSearch)
	r.Get("/rss", app.handleRSS)
	fileServer(r, "/static", http.Dir("static"))

	// Catch-all handler for 404s
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		app.render404(w, r)
	})

	log.Printf("[server] listening on %s ...", *port)
	if err := http.ListenAndServe(*port, r); err != nil {
		log.Fatal("[server] fatal:", err)
	}
}

func NewApp(postsDir string, perPage int, siteTitle string) (*App, error) {
	tpl, err := template.ParseFiles(
		filepath.Join("templates", "layout.html"),
		filepath.Join("templates", "index.html"),
		filepath.Join("templates", "post.html"),
		filepath.Join("templates", "404.html"),
	)

	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	app := &App{
		PerPage:   perPage,
		Templates: tpl,
		SiteTitle: siteTitle,
		PostsDir:  postsDir,
		Markdown:  goldmark.New(),
	}

	if err := app.reloadPosts(); err != nil {
		return nil, err
	}
	log.Printf("[init] loaded %d posts", len(app.Posts))

	go app.watchPosts()

	return app, nil
}

// domain utility
func loadPosts(dir string) ([]PostMeta, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read posts dir: %w", err)
	}

	var posts []PostMeta
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		m := filenameRe.FindStringSubmatch(name)
		if m == nil {
			log.Printf("[load] skip invalid filename: %s", name)
			continue
		}
		ts := m[1]
		slug := m[2]

		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			log.Printf("[load] read error: %s: %v", name, err)
			continue
		}

		raw := string(b)
		title, tags, _ := parseMeta(raw)
		if title == "" {
			title = deriveTitleFromSlug(slug)
		}
		t, err := parseTimestamp(ts)
		if err != nil {
			log.Printf("[load] bad timestamp: %s (%s)", ts, name)
			continue
		}
		posts = append(posts, PostMeta{Title: title, Slug: slug, Date: t, Tags: tags, Filename: name})
		log.Printf("[load] loaded: %s (%s)", title, name)

	}
	return posts, nil
}

func parseMeta(raw string) (title string, tags []string, body string) {
	lines := strings.Split(raw, "\n")
	contentStart := 0
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			contentStart = i + 1
			break
		}
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
		} else if strings.HasPrefix(line, "tag:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "tag:"))
			parts := strings.Split(tagStr, ",")
			for _, p := range parts {
				t := strings.ToLower(strings.TrimSpace(p))
				if t != "" {
					tags = append(tags, t)
				}
			}
		}
	}
	body = strings.Join(lines[contentStart:], "\n")
	return
}

func parseTimestamp(s string) (time.Time, error) {
	switch len(s) {
	case 8:
		return time.ParseInLocation(dateLayoutDay, s, time.Local)
	case 14:
		return time.ParseInLocation(dateLayoutSecond, s, time.Local)
	default:
		return time.Time{}, errors.New("invalid timestamp length")
	}
}

func deriveTitleFromSlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func (a *App) render404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := map[string]any{
		"Title": "Page Not Found · " + a.SiteTitle,
		"Is404": true,
	}
	if err := a.Templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("[404] render error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func fileServer(r *chi.Mux, path string, root http.FileSystem) {
	fs := http.StripPrefix(path, http.FileServer(root))
	r.Get(path+"/*", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

//  handler

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	a.renderIndex(w, 1, r)
}

func (a *App) handleIndexPage(w http.ResponseWriter, r *http.Request) {
	nStr := chi.URLParam(r, "n")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 {
		log.Printf("[index] invalid page param: %s", nStr)
		a.render404(w, r)
		return
	}
	a.renderIndex(w, n, r)
}

func (a *App) renderIndex(w http.ResponseWriter, page int, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := len(a.Posts)
	start := (page - 1) * a.PerPage
	if start >= total && page != 1 {
		log.Printf("[index] page out of range: %d", page)
		a.render404(w, r)
		return
	}
	end := start + a.PerPage
	if end > total {
		end = total
	}
	items := a.Posts[start:end]

	data := map[string]any{
		"Title":   a.SiteTitle,
		"Posts":   items,
		"Page":    page,
		"HasPrev": page > 1,
		"HasNext": end < total,
		"PrevURL": fmt.Sprintf("/page/%d", page-1),
		"NextURL": fmt.Sprintf("/page/%d", page+1),
		"IsFirst": page == 1,
	}
	if page == 1 {
		data["PrevURL"] = "/"
	}
	if err := a.Templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("[index] render error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handlePost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	a.mu.RLock()
	idx, ok := a.BySlug[slug]
	if !ok {
		a.mu.RUnlock()
		log.Printf("[post] not found: %s", slug)
		a.render404(w, r)
		return
	}
	meta := a.Posts[idx]
	a.mu.RUnlock()

	b, err := os.ReadFile(filepath.Join(a.PostsDir, meta.Filename))
	if err != nil {
		log.Printf("[post] read error: %s: %v", meta.Filename, err)
		http.Error(w, "cannot read post", http.StatusInternalServerError)
		return
	}
	_, _, body := parseMeta(string(b))

	var buf strings.Builder
	if err := a.Markdown.Convert([]byte(body), &buf); err != nil {
		log.Printf("[post] markdown error: %s: %v", meta.Filename, err)
		http.Error(w, "markdown error", http.StatusInternalServerError)
		return
	}

	post := Post{Meta: meta, HTML: template.HTML(buf.String())}
	data := map[string]any{"Title": a.SiteTitle + " · " + meta.Title, "Post": post, "IsPost": true}
	if err := a.Templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("[post] render error: %s: %v", slug, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleTag(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")

	a.mu.RLock()
	idxs, ok := a.TagIndex[tag]
	if !ok || len(idxs) == 0 {
		a.mu.RUnlock()
		log.Printf("[tag] not found: %s", tag)
		a.render404(w, r)
		return
	}
	var items []PostMeta
	for _, i := range idxs {
		items = append(items, a.Posts[i])
	}
	a.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool { return items[i].Date.After(items[j].Date) })
	data := map[string]any{"Title": fmt.Sprintf("Tag: %s · %s", tag, a.SiteTitle), "Posts": items, "Tag": tag}
	if err := a.Templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("[tag] render error: %s: %v", tag, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if q == "" {
		log.Printf("[search] empty query")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	a.mu.RLock()
	var results []PostMeta
	for _, p := range a.Posts {
		if strings.Contains(strings.ToLower(p.Title), q) {
			results = append(results, p)
			continue
		}
		for _, tag := range p.Tags {
			if strings.Contains(tag, q) {
				results = append(results, p)
				break
			}
		}
	}
	a.mu.RUnlock()

	if len(results) == 0 {
		log.Printf("[search] no results for query: %s", q)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Date.After(results[j].Date) })
	data := map[string]any{"Title": fmt.Sprintf("Search: %s", q), "Posts": results, "Query": q, "IsSearch": true}
	if err := a.Templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("[search] render error: %s: %v", q, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleRSS(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	posts := make([]PostMeta, len(a.Posts))
	copy(posts, a.Posts)
	a.mu.RUnlock()

	// Limit to recent 20 posts for RSS
	if len(posts) > 20 {
		posts = posts[:20]
	}

	// Set content type for RSS/XML
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

	// Build RSS XML manually
	rssContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>%s</title>
    <description>Latest posts from %s</description>
    <link>%s</link>
    <atom:link href="%s/rss" rel="self" type="application/rss+xml"/>
    <lastBuildDate>%s</lastBuildDate>
    <language>en-us</language>`,
		a.SiteTitle,
		a.SiteTitle,
		getBaseURL(r),
		getBaseURL(r),
		time.Now().Format(time.RFC1123Z))

	for _, post := range posts {
		// Read post content for description
		description := post.Title
		if content := a.getPostContent(post); content != "" {
			// Truncate content for description (first 200 chars)
			if len(content) > 200 {
				description = content[:200] + "..."
			} else {
				description = content
			}
		}

		// Escape XML content
		description = template.HTMLEscapeString(description)
		title := template.HTMLEscapeString(post.Title)

		rssContent += fmt.Sprintf(`
    <item>
      <title>%s</title>
      <description>%s</description>
      <link>%s/post/%s</link>
      <guid>%s/post/%s</guid>
      <pubDate>%s</pubDate>`,
			title,
			description,
			getBaseURL(r),
			post.Slug,
			getBaseURL(r),
			post.Slug,
			post.Date.Format(time.RFC1123Z))

		// Add categories (tags)
		for _, tag := range post.Tags {
			rssContent += fmt.Sprintf(`
      <category>%s</category>`, template.HTMLEscapeString(tag))
		}

		rssContent += `
    </item>`
	}

	rssContent += `
  </channel>
</rss>`

	w.Write([]byte(rssContent))
}

func (a *App) getPostContent(meta PostMeta) string {
	b, err := os.ReadFile(filepath.Join(a.PostsDir, meta.Filename))
	if err != nil {
		log.Printf("failed read file: %v", err)
		return ""
	}
	_, _, body := parseMeta(string(b))

	// Remove markdown formatting for RSS description
	lines := strings.Split(body, "\n")
	var plainText []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") {
			continue
		}
		// Remove basic markdown formatting
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "*", "")
		line = strings.ReplaceAll(line, "`", "")
		if line != "" {
			plainText = append(plainText, line)
		}
		if len(strings.Join(plainText, " ")) > 300 {
			break
		}
	}
	return strings.Join(plainText, " ")
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
