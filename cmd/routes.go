package main

import (
	"backstreetlinkv2/cmd/middleware"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func(app *Application) routes() http.Handler {
	mux := mux.NewRouter()

	c := cors.New(cors.Options{
        AllowedOrigins: []string{"https://backstreet.link", "https://www.backstreet.link"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Authorization"},
        Debug: false,
    })

	mux.Use(
		middleware.Recoverer,
	)

	app.createLinkRoutes(mux)
	app.createFileRoutes(mux)
	app.findDataRoutes(mux)
	app.requestDownloadFileRoutes(mux)
	
	return c.Handler(mux) 
}

func(app *Application) createLinkRoutes(mux *mux.Router) {
	link := mux.PathPrefix("/api/v2").Subrouter()
	link.HandleFunc("/link", app.Create).Methods(http.MethodPost)
	link.Use(middleware.OnlyJSONRequest)
}

func(app *Application) createFileRoutes(mux *mux.Router) {
	file := mux.PathPrefix("/api/v2").Subrouter()
	file.HandleFunc("/file", app.CreateMultipart).Methods(http.MethodPost)
	file.Use(middleware.SetResponseHeaderJSON)
}

func(app *Application) requestDownloadFileRoutes(mux *mux.Router) {
	downloadFile := mux.PathPrefix("/api/v2").Subrouter()
	downloadFile.HandleFunc("/download-file/{alias}", app.DownloadFile).Methods(http.MethodGet)
}

func(app *Application) findDataRoutes(mux *mux.Router) {
	findData := mux.PathPrefix("/api/v2").Subrouter()
	findData.HandleFunc("/find/{alias}", app.Find).Methods(http.MethodGet)
	findData.Use(middleware.SetResponseHeaderJSON)
}