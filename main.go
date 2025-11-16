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
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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
	mux.HandleFunc("/blog", s.handleList(blogFS, "blogs", "blogs"))
	mux.HandleFunc("/wiki", s.handleList(wikiFS, "wikis", "wikis"))
	mux.HandleFunc("/blog/", s.handleContent(blogFS, "blogs"))
	mux.HandleFunc("/wiki/", s.handleContent(wikiFS, "wikis"))

	return mux
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", loggingMiddleware(srv.routes())); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := ViewData{
		Title: "Home page",
	}

	templateName := "layout"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "home"
	}

	if err := s.templates.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) handleList(fsys embed.FS, dir string, listType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := listFiles(fsys, dir)
		if err != nil {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
		}

		data := ViewData{}
		if listType == "blogs" {
			data.Blogs = files
		} else {
			data.Wikis = files
		}

		templateName := "layout"
		if r.Header.Get("HX-Request") == "true" {
			templateName = "index"
		}

		if err := s.templates.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *server) handleContent(contentFS embed.FS, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := path.Base(r.URL.Path)

		if path.Ext(fileName) != "" {
			filePath := path.Join(contentType, fileName)
			content, err := fs.ReadFile(contentFS, filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			info, err := fs.Stat(contentFS, filePath)
			if err != nil {
				http.Error(w, "Failed to get file info", http.StatusInternalServerError)
				return
			}

			http.ServeContent(w, r, fileName, info.ModTime(), bytes.NewReader(content))
			return
		}

		filePath := path.Join(contentType, fileName+".md")

		markdown, err := fs.ReadFile(contentFS, filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var buf bytes.Buffer
		md := goldmark.New(
			goldmark.WithParserOptions(
				parser.WithASTTransformers(
					util.Prioritized(&linkTransformer{}, 100),
				),
			),
		)
		if err := md.Convert(markdown, &buf); err != nil {
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

type linkTransformer struct {
}

func (t *linkTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if link, ok := n.(*ast.Link); ok {
			dest := string(link.Destination)
			if strings.HasSuffix(dest, ".md") && !strings.HasPrefix(dest, "http") {
				if strings.Contains(dest, "wikis/") {
					base := path.Base(dest)
					fileName := strings.TrimSuffix(base, ".md")
					link.Destination = []byte("/wiki/" + fileName)
				} else if strings.Contains(dest, "blogs/") {
					base := path.Base(dest)
					fileName := strings.TrimSuffix(base, ".md")
					link.Destination = []byte("/blog/" + fileName)
				}
			}
		}
		return ast.WalkContinue, nil
	})
}
