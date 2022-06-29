package main

import (
	"backstreetlinkv2/db"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
)

// Application struct untuk menghandle core server
type Application struct {
	Server               *http.Server
	Connection           *mongo.Client
	InfoLog              *log.Logger
	ErrorLog             *log.Logger
	IdleConnectionClosed chan struct{}
}

func main() {
	client, err := db.Connect(os.Getenv("MONGODB_URI"))
	if err != nil {
		log.Fatal(err)
	}

	app := Application{
		Connection:           client,
		InfoLog:              log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:             log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		IdleConnectionClosed: make(chan struct{}),
	}

	go app.shutdown()

	app.serve()

	<-app.IdleConnectionClosed
	app.InfoLog.Println("App stopped successfully!!")
}

// serve menangani bagian saat mau start server
func (app *Application) serve() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.Server = &http.Server{
		Addr:              ":"+port,
		Handler:           app.routes(),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    10 * 1024 * 1024,
	}

	app.InfoLog.Println("Server start...")

	if err := app.Server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			app.ErrorLog.Fatal("Server failed to start:", err)
		}
	}
}

// shutdown menangani bagian saat mau nutup server
func (app *Application) shutdown() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	<-sigint

	app.InfoLog.Println("Shutdown order received!")

	//closed db connection
	closedDBChan := make(chan bool, 1)
	go app.closeDB(closedDBChan)

	//shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		app.ErrorLog.Printf("Server shutdown error: %v", err)
	}

	<-closedDBChan

	app.InfoLog.Println("Shutdown complete")

	close(app.IdleConnectionClosed)
	close(closedDBChan)
}

// closeDB menutup koneksi ke database
func (app *Application) closeDB(done chan bool) {
	app.InfoLog.Println("Closing database connection")

	if err := app.Connection.Disconnect(context.TODO()); err != nil {
		app.ErrorLog.Fatal(err)
	}

	done <- true
}
