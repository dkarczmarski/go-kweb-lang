package web

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type TemplateData struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewTemplateData() *TemplateData {
	return &TemplateData{
		data: make(map[string]any),
	}
}

func (td *TemplateData) Set(key string, data any) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.data[key] = data
}

func (td *TemplateData) Get(key string) any {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.data[key]
}

type Server struct {
	httpServer *http.Server
}

//go:embed index.html
var indexHTML string

func NewServer(templateData *TemplateData) *Server {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}
	tmpl := template.Must(template.New("index.html").Funcs(funcMap).Parse(indexHTML))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /")

		model := templateData.Get("pl")
		if err := tmpl.Execute(w, model); err != nil {
			log.Fatal(err)
		}
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return &Server{
		httpServer: httpServer,
	}
}

func (srv *Server) ListenAndServe() error {
	return srv.httpServer.ListenAndServe()
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}
