package main

import (
	help "backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"backstreetlinkv2/cmd/repo"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	ErrInternal = http.StatusInternalServerError
	BadRequest = http.StatusBadRequest
)

func TestHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello World"))
}

func (app *Application) Create(w http.ResponseWriter, r *http.Request) {
	data, err := decodeShortenRequest(r)
	if err != nil {
		w.WriteHeader(BadRequest)
		json.NewEncoder(w).Encode(help.WriteErr(err))
		return
	}

	err = repo.InsertLink(r.Context(), app.Connection, data)
	if err != nil {
		w.WriteHeader(BadRequest)
		json.NewEncoder(w).Encode(help.DBErrResponse(err))
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), ErrInternal)
		return
	}
}

func (app *Application) Find(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	data, err := repo.Find(r.Context(), app.Connection, vars["alias"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(help.DBErrResponse(err))
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(ErrInternal)
		json.NewEncoder(w).Encode(help.WriteErr(err))
		return
	}
}

func (app *Application) DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	data, err := repo.Find(r.Context(), app.Connection, vars["alias"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(help.DBErrResponse(err))
		return
	}

	if data.Type == model.TypeFile {
		blob, err := repo.HandleFindFileRequest(r.Context(), data)
		if err != nil {
			w.WriteHeader(ErrInternal)
			json.NewEncoder(w).Encode(help.WriteErr(err))
			return
		}

		defer blob.Body.Close()

		disposition := fmt.Sprintf("attachment; filename=\"%s\"", data.Filename)
		contentType := *blob.ContentType
		contentLength := strconv.FormatInt(*blob.ContentLength, 10)

		w.Header().Set("Content-Disposition", disposition)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", contentLength)

		io.Copy(w, blob.Body)
		return
	}

	w.WriteHeader(ErrInternal)
	json.NewEncoder(w).Encode(help.WriteErr(errors.New("can't process this request right now")))
}

func decodeShortenRequest(r *http.Request) (model.ShortenRequest, error) {
	var shorten model.ShortenRequest
	
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&shorten); err != nil {
		return shorten, err
	}

	if err := help.ValidateStruct(shorten); err != nil {
		return shorten, err
	}

	return shorten, nil
}
