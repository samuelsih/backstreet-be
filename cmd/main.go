package main

import (
	"backstreetlinkv2/cmd/middleware"
	"backstreetlinkv2/cmd/repo"
	"backstreetlinkv2/cmd/service"
	"backstreetlinkv2/db"
	"backstreetlinkv2/db/migrations"
	"context"
	"errors"
	"flag"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 30 * time.Second

func main() {
	wantFreshDB := flag.Bool("fresh", false, "drop the DB and remigrate it")
	flag.Parse()

	log.SetOutput(zerolog.New(os.Stdout))

	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "LOCAL"
	}

	dsn := os.Getenv("COCKROACH_DSN")
	if dsn == "" {
		log.Fatal("no dsn")
	}

	dbClient, err := db.ConnectPG(dsn)
	if err != nil {
		log.Fatalf("can't connect to db: %v", err)
	}

	if *wantFreshDB {
		ctx := context.Background()

		_, err := dbClient.Exec(ctx, migrations.DownCmd)
		if err != nil {
			log.Fatalf("cant drop table: %v", err)
		}

		_, err = dbClient.Exec(ctx, migrations.UpCmd)
		if err != nil {
			log.Fatalf("cant create table: %v", err)
		}
	}

	port := os.Getenv("port")
	if port == "" {
		port = ":8080"
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("AWS_BUCKETNAME")

	router := mux.NewRouter()
	router.Use(
		middleware.CORS(environment),
		middleware.Recoverer,
		middleware.Limit,
	)

	pgRepo := repo.NewPGRepo(dbClient)
	s3Service, err := repo.NewObjectScanner(repo.ObjectConfig{
		AccessKey:        accessKey,
		SecretKey:        secretKey,
		Endpoint:         endpoint,
		Region:           region,
		ForceS3PathStyle: false,
		Bucket:           bucket,
	})

	if err != nil {
		log.Fatalf("error s3: %v", err)
	}

	programService := service.NewLinkDeps(pgRepo, s3Service)

	r := router.PathPrefix("/api/v2").Subrouter()
	r.HandleFunc("/link", createLink(programService)).Methods(http.MethodPost)
	r.HandleFunc("/file", createFile(programService)).Methods(http.MethodPost)
	r.HandleFunc("/download-file/{alias}", downloadFile(programService)).Methods(http.MethodGet)
	r.HandleFunc("/find/{alias}", find(programService)).Methods(http.MethodGet)

	server := &http.Server{
		Addr:              ":" + port,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
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

func getFromEnvWithDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
}
