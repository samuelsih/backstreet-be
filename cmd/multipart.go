package main

import (
	help "backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"backstreetlinkv2/cmd/repo"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

var (
	ErrCantProcess        = errors.New("can't process request right now. try again later")
	ErrMaxSize            = errors.New("file size over limit, max is 32MB")
	ErrUnexpectedBehavior = errors.New("unexpected behavior")
)

const (
	maxSize int64 = 32 << 20 //32MB
)

func (app *Application) CreateMultipart(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	reader, err := r.MultipartReader()
	if err != nil {
		app.ErrorLog.Println(err)
		w.WriteHeader(BadRequest)
		json.NewEncoder(w).Encode(help.WriteErr(ErrCantProcess))
		return
	}

	var jsonRequest model.ShortenFileRequest
	var fileRequest *multipart.Part

	for i := 0; i < 2; i++ {
		part, err := reader.NextPart()
		if (err == io.EOF) && (i <= 1) {
			app.ErrorLog.Println(err)
			w.WriteHeader(BadRequest)
			json.NewEncoder(w).Encode(help.WriteErr(ErrUnexpectedBehavior))
			return
		}

		switch part.FormName() {
		case "json_field":
			jsonRequest, err = decodeMultipartRequest(part)
			if err != nil {
				app.ErrorLog.Println(err)
				w.WriteHeader(BadRequest)
				json.NewEncoder(w).Encode(help.WriteErr(ErrCantProcess))
				return
			}

		case "file_field":
			if err := app.checkLimitFileSize(part); err != nil {
				app.ErrorLog.Println(err)
				w.WriteHeader(BadRequest)
				json.NewEncoder(w).Encode(help.WriteErr(err))
				return
			}

			fileRequest = part

		default:
			err := fmt.Sprintf("unexpected field %v", part.FormName())
			app.ErrorLog.Println(err)

			w.WriteHeader(BadRequest)
			json.NewEncoder(w).Encode(help.WriteErr(errors.New(err)))
			return
		}
	}

	result, err := repo.HandleS3(r.Context(), app.Connection, jsonRequest, fileRequest)
	if err != nil {
		app.ErrorLog.Println(err)
		w.WriteHeader(ErrInternal)
		json.NewEncoder(w).Encode(help.WriteErr(err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(result)
}

func decodeMultipartRequest(r *multipart.Part) (model.ShortenFileRequest, error) {
	var request model.ShortenFileRequest

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		return request, err
	}

	if err := help.ValidateStruct(request); err != nil {
		return request, err
	}

	return request, nil
}

func (app *Application) checkLimitFileSize(part *multipart.Part) error {
	buf := bufio.NewReader(part)

	file, err := os.CreateTemp("", "")
	if err != nil {
		app.ErrorLog.Println(err)
		return ErrCantProcess
	}

	defer file.Close()

	limit := io.MultiReader(buf, io.LimitReader(part, maxSize-511))
	written, err := io.Copy(file, limit)

	if err != nil || written > maxSize {
		app.ErrorLog.Println(err)
		os.Remove(file.Name())
		return ErrMaxSize
	}

	return nil
}
