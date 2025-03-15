package web

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

//go:embed index.html
var indexHTML string

//go:embed lang.html
var langHTML string

func NewServer(templateData *TemplateData) *Server {
	funcMap := template.FuncMap{
		"truncate": truncate,
	}
	indexTmpl := template.Must(template.New("index.html").Funcs(funcMap).Parse(indexHTML))
	langTmpl := template.Must(template.New("lang.html").Funcs(funcMap).Parse(langHTML))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		model := templateData.GetIndex()
		if err := indexTmpl.Execute(w, model); err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/lang/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")

		model := templateData.GetLang(code)
		if model == nil {
			http.NotFound(w, r)
			return
		}

		if err := langTmpl.Execute(w, model); err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
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
