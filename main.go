package main

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/yuin/goldmark"
)

//go:embed blogs
var blogFS embed.FS

//go:embed wikis
var wikiFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed templates
var templateFS embed.FS

type ViewData struct {
	Title   string
	Content template.HTML
	Blogs []string
	Wikis []string
}

type server struct {
	templates *template.Template
}

func newServer() (*server, error) {
	templates, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &server{
		templates: templates,
	}, nil
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()

	staticServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/static/", staticServer)

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/blog/", s.handleContent(blogFS, "blogs"))
	mux.HandleFunc("/wiki/", s.handleContent(wikiFS, "wikis"))

	return mux
}

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", srv.routes()); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	blogs, err := listFiles(blogFS, "blogs")
	if err != nil {
		http.Error(w, "Failed to list blogs", http.StatusInternalServerError)
		return
	}

	wikis, err := listFiles(wikiFS, "wikis")
	if err != nil {
		http.Error(w, "Failed to list wikis", http.StatusInternalServerError)
		return
	}

	data := ViewData{
		Blogs: blogs,
		Wikis: wikis,
	}

	templateName := "layout"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "index"
	}

	if err := s.templates.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) handleContent(contentFS embed.FS, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := path.Base(r.URL.Path)
		filePath := path.Join(contentType, fileName+".md")

		markdown, err := fs.ReadFile(contentFS, filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var buf bytes.Buffer
		if err := goldmark.Convert(markdown, &buf); err != nil {
			http.Error(w, "Failed to render markdown", http.StatusInternalServerError)
			return
		}

		data := ViewData{
			Title:   strings.ReplaceAll(fileName, "-", " "),
			Content: template.HTML(buf.String()),
		}

		templateName := "content"
		if r.Header.Get("HX-Request") != "true" {
			templateName = "layout"
		}

		if err := s.templates.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func listFiles(fs embed.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := strings.TrimSuffix(entry.Name(), ".md")
			files = append(files, name)
		}
	}
	return files, nil
}
