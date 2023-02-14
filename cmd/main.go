package main

import (
	"backstreetlinkv2/cmd/middleware"
	"backstreetlinkv2/db"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 30 * time.Second

func main() {
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "LOCAL"
	}

	dsn := os.Getenv("COCKROACH_DSN")
	if dsn == "" {
		log.Fatal("no dsn cockroach")
	}

	dbClient, err := db.ConnectPG(dsn)
	if err != nil {
		log.Fatalf("can't connect to db: %v", err)
	}

	port := os.Getenv("port")
	if port == "" {
		port = ":8080"
	}

	mux := mux.NewRouter()
	mux.Use(
		middleware.CORS(environment),
		middleware.Recoverer,
		middleware.Limit,
	)

	r := mux.PathPrefix("/api/v2").Subrouter()
	r.HandleFunc("/link", createLink()).Methods(http.MethodPost)
	r.HandleFunc("/file", CreateFile()).Methods(http.MethodPost)
	r.HandleFunc("/download-file/{alias}", downloadFile()).Methods(http.MethodGet)
	r.HandleFunc("/find/{alias}", find()).Methods(http.MethodGet)

	server := &http.Server{
		Addr:              ":" + port,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    10 * 1024 * 1024,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("cant start server: %v", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	<-quit

	// for this case, imho errgroup / goroutine is overkill
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := dbClient.Close(ctx); err != nil {
		log.Printf("shutting down db error: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutting down server error: %v", err)
	}
}
