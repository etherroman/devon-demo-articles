package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"devon-demo-articles/internal/application"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	if os.Getenv("GIN_MODE") == "" {
		_ = os.Setenv("GIN_MODE", "release")
	}
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "data/articles.db"
	}
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = "127.0.0.1:8082"
	}
	app, err := application.New(path, os.Getenv("JWT_SECRET"))
	if err != nil {
		return err
	}
	defer app.Close()
	server := &http.Server{Addr: address, Handler: app.Handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1024 * 1024}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	result := make(chan error, 1)
	go func() { log.Printf("articles API listening on http://%s", address); result <- server.ListenAndServe() }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
