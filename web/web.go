package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dkarczmarski/go-kweb-lang/dashboard"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	maxHeaderBytes    = 50 * 1024
)

type Server struct {
	httpServer *http.Server
}

func NewServer(webHTTPAddr string, dashboardStore *dashboard.Store) *Server {
	mux := http.NewServeMux()
	handler := NewHandler(dashboardStore)
	handler.Register(mux)

	//nolint:exhaustruct
	httpServer := &http.Server{
		Addr:              webHTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	return &Server{
		httpServer: httpServer,
	}
}

func (server *Server) ListenAndServe() error {
	err := server.httpServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("listen and serve web server: %w", err)
	}

	return nil
}

func (server *Server) Shutdown(ctx context.Context) error {
	err := server.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("shutdown web server: %w", err)
	}

	return nil
}
