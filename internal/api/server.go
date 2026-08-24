package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(address string, handler http.Handler) *HTTPServer {
	return &HTTPServer{server: &http.Server{
		Addr: address, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}}
}

func (s *HTTPServer) Serve(listener net.Listener) error {
	err := s.server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
