package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/cry-109/internal/api"
	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/timing"
)

func main() {
	address := flag.String("addr", envOr("PORTCRANE_ADDR", "127.0.0.1:8080"), "HTTP listen address")
	dataDir := flag.String("data", envOr("PORTCRANE_DATA", "./data"), "local state directory")
	flag.Parse()

	runtime, err := control.NewRuntime(*dataDir, timing.RealClock{})
	if err != nil {
		slog.Error("initialize runtime", "error", err)
		os.Exit(1)
	}
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		slog.Error("listen", "address", *address, "error", err)
		os.Exit(1)
	}
	server := api.NewHTTPServer(*address, api.NewServer(runtime).Handler())
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()
	slog.Info("portcrane ready", "address", listener.Addr().String(), "data", *dataDir)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case signal := <-signals:
		slog.Info("shutdown requested", "signal", signal.String())
	case err := <-serveResult:
		if err != nil {
			slog.Error("http server stopped", "error", err)
			os.Exit(1)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown", "error", err)
		os.Exit(1)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
