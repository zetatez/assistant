package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zetatez/assistant/internal/config"
	"github.com/zetatez/assistant/internal/db"
	"github.com/zetatez/assistant/internal/http"
)

func main() {
	cfg := config.Load()

	// init DB
	gormDB, err := db.New(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer func() { sqlDB, _ := gormDB.DB(); _ = sqlDB.Close() }()

	r := httpserver.NewRouter(cfg, gormDB)

	srv := &http.Server{
		Addr:         cfg.App.Addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("%s running on %s...", cfg.App.Name, cfg.App.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exiting")
}
