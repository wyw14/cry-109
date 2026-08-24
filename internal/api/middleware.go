package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *Server) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		requestID := request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		writer.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(writer, request)
		slog.Debug("request", "id", requestID, "method", request.Method, "path", request.URL.Path, "duration", time.Since(started))
	})
}

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("request panic", "path", request.URL.Path, "error", recovered)
				writeError(writer, http.StatusInternalServerError, "internal control error")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
