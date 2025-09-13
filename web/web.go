package web

import (
	"context"
	"net/http"
	"time"

	"go-kweb-lang/dashboard"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(webHTTPAddr string, dashboardStore *dashboard.Store) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", createListLangCodesHandler(dashboardStore))
	mux.HandleFunc("GET /lang/{code}", createLangDashboardHandler(dashboardStore))
	mux.HandleFunc("POST /lang/{code}", createLangDashboardTableHandler(dashboardStore))

	httpServer := &http.Server{
		Addr:              webHTTPAddr,
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
