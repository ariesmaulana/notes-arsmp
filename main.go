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
	dateLayoutSecond = "20060102150505"
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
	postDir := flag.String("posts", defaultPostsDir, "Directory containing post")
	port := flag.String("port", defaultPort, "HTTP listen address")
	perPage := flag.Int("perpage", defaultPerPage, "Post per page")
	siteTitle := flag.String("title", "Arsmp", "Site Title")
	flag.Parse()

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
