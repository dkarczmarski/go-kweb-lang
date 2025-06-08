package web

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(viewModelStore ViewModelStore) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", createListLangCodesHandler(viewModelStore))
	mux.HandleFunc("GET /lang/{code}", createLangDashboardHandler(viewModelStore))
	mux.HandleFunc("POST /lang/{code}", createLangDashboardTableHandler(viewModelStore))

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    50 * 1024,
	}

	return &Server{
		httpServer: httpServer,
	}
}

func (srv *Server) ListenAndServe() error {
	return srv.httpServer.ListenAndServe()
}

func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.httpServer.Shutdown(ctx)
}
