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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

//go:embed blogs
var blogFS embed.FS

//go:embed wikis
var wikiFS embed.FS

//go:embed assets
var assetsFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed templates
var templateFS embed.FS

type ViewData struct {
	Title       string
	Content     template.HTML
	Blogs       []ContentItem
	Wikis       []ContentItem
	TOC         []Heading
	Description string
}

type ContentItem struct {
	FileName string
	Title    string
}

type Heading struct {
	Level int
	Text  string
	ID    string
}

type FrontMatter struct {
	Title string `yaml:"title"`
	Desc  string `yaml:"desc"`
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

	assetsServer := http.FileServer(http.FS(assetsFS))
	mux.Handle("/assets/", assetsServer)

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

	var port string = ":12352"

	log.Println("Server starting on port", port)
	if err := http.ListenAndServe(port, loggingMiddleware(srv.routes())); err != nil {
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
		items, err := listContentItems(fsys, dir)
		if err != nil {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
		}

		data := ViewData{}
		if listType == "blogs" {
			data.Title = "Blogs"
			data.Blogs = items
		} else {
			data.Title = "Wikis"
			data.Wikis = items
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

func extractFrontMatter(markdown []byte) (FrontMatter, []byte, error) {
	var fm FrontMatter
	content := string(markdown)
	parts := strings.SplitN(content, "---", 3)

	if len(parts) < 3 || strings.TrimSpace(parts[0]) != "" {
		return fm, markdown, nil
	}

	yamlBlock := strings.TrimSpace(parts[1])
	markdownContent := []byte(strings.TrimSpace(parts[2]))

	if yamlBlock == "" {
		return fm, markdown, nil
	}

	return fm, markdownContent, yaml.Unmarshal([]byte(yamlBlock), &fm)
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

		fm, markdownContent, err := extractFrontMatter(markdown)
		if err != nil {
			log.Printf("Error parsing front matter for %s: %v", filePath, err)
		}

		title := strings.ReplaceAll(fileName, "-", " ")
		if fm.Title != "" {
			title = fm.Title
		}

		var headings []Heading
		var buf bytes.Buffer
		md := goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
			),
			goldmark.WithParserOptions(
				parser.WithASTTransformers(
					util.Prioritized(&linkTransformer{}, 100),
					util.Prioritized(&tocExtractor{Headings: &headings}, 200),
				),
			),
		)
		if err := md.Convert(markdownContent, &buf); err != nil {
			http.Error(w, "Failed to render markdown", http.StatusInternalServerError)
			return
		}

		data := ViewData{
			Title:       title,
			Content:     template.HTML(buf.String()),
			TOC:         headings,
			Description: fm.Desc,
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

func listContentItems(fsys embed.FS, dir string) ([]ContentItem, error) {
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []ContentItem
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		fileName := strings.TrimSuffix(entry.Name(), ".md")
		filePath := path.Join(dir, entry.Name())

		markdown, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v", filePath, err)
			continue
		}

		fm, _, err := extractFrontMatter(markdown)
		if err != nil {
			log.Printf("Error parsing front matter for %s: %v", filePath, err)
		}

		title := fileName
		if fm.Title != "" {
			title = fm.Title
		}

		items = append(items, ContentItem{
			FileName: fileName,
			Title:    title,
		})
	}
	return items, nil
}

type tocExtractor struct {
	Headings *[]Heading
}

func (t *tocExtractor) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok {
			idStr := ""
			if idAttr, ok := heading.AttributeString("id"); ok {
				if idBytes, ok := idAttr.([]byte); ok {
					idStr = string(idBytes)
				}
			}
			*t.Headings = append(*t.Headings, Heading{
				Level: heading.Level,
				Text:  string(heading.Text(reader.Source())),
				ID:    idStr,
			})
		}
		return ast.WalkContinue, nil
	})
}

type linkTransformer struct {
}

func (t *linkTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Link:
			dest := string(v.Destination)
			if strings.HasSuffix(dest, ".md") && !strings.HasPrefix(dest, "http") {
				if strings.Contains(dest, "wikis/") {
					base := path.Base(dest)
					fileName := strings.TrimSuffix(base, ".md")
					v.Destination = []byte("/wiki/" + fileName)
				} else if strings.Contains(dest, "blogs/") {
					base := path.Base(dest)
					fileName := strings.TrimSuffix(base, ".md")
					v.Destination = []byte("/blog/" + fileName)
				}
			}
		case *ast.Image:
			dest := string(v.Destination)
			if !strings.HasPrefix(dest, "http") && !path.IsAbs(dest) {
				v.Destination = []byte(path.Join("/", dest))
			}
		}

		return ast.WalkContinue, nil
	})
}
