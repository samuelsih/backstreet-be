package main

import (
	"backstreetlinkv2/api"
	"backstreetlinkv2/api/middleware"
	"backstreetlinkv2/api/repo"
	"backstreetlinkv2/api/service"
	"backstreetlinkv2/db"
	"backstreetlinkv2/db/migrations"
	"context"
	"errors"
	"flag"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 30 * time.Second
const maxCacheCountdown = 12 * time.Hour

func main() {
	wantFreshDB := flag.Bool("fresh", false, "drop the DB and remigrate it")
	flag.Parse()

	log.SetOutput(zerolog.New(os.Stdout))

	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "LOCAL"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/backstreet"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	accessKey := os.Getenv("NEW_AWS_ACCESS_KEY")
	secretKey := os.Getenv("NEW_AWS_SECRET_KEY")
	endpoint := os.Getenv("NEW_AWS_ENDPOINT")
	bucket := os.Getenv("NEW_AWS_BUCKETNAME")

	router := mux.NewRouter()
	router.Use(
		middleware.CORS(environment),
		middleware.Recoverer,
		middleware.Limit,
	)

	pgRepo := repo.NewPGRepo(dbClient)
	cache, err := repo.NewCache(context.Background(), maxCacheCountdown)
	if err != nil {
		log.Fatalf("error cache: %v", err)
	}

	s3Service, err := repo.NewObjectScanner(context.Background(), repo.ObjectConfig{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Endpoint:  endpoint,
		Bucket:    bucket,
	})

	if err != nil {
		log.Fatalf("error s3: %v", err)
	}

	programService := service.NewLinkDeps(pgRepo, s3Service, cache)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	r := router.PathPrefix("/api/v2").Subrouter()
	r.HandleFunc("/link", api.CreateLink(programService)).Methods(http.MethodPost)
	r.HandleFunc("/file", api.CreateFile(programService)).Methods(http.MethodPost)
	r.HandleFunc("/download-file/{alias}", api.DownloadFile(programService)).Methods(http.MethodGet)
	r.HandleFunc("/find/{alias}", api.Find(programService)).Methods(http.MethodGet)

	server := &http.Server{
		Addr:              ":" + port,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    10 * 1024 * 1024,
		Handler:           router,
	}

	go func() {
		log.Println("currently listen and serve on port", port)

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
