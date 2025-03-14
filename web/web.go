package web

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

type TemplateData struct {
	// todo: sync
	data any
}

func (td *TemplateData) Set(data any) {
	td.data = data
}

func (td *TemplateData) Get() any {
	return td.data
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

		model := templateData.Get()
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
