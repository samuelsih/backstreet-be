package main

import (
	"backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"backstreetlinkv2/cmd/service"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	statusBadReq      = http.StatusBadRequest
	statusNotFound    = http.StatusNotFound
	statusInternalErr = http.StatusInternalServerError
)

func createLink(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request model.ShortenRequest

		err := decodeJSONLinkRequest(r, &request)
		if err != nil {
			w.WriteHeader(statusNotFound)
			sendJSONErr(w, statusNotFound, err.Error())
			return
		}

		output := svc.InsertLink(r.Context(), request)
		w.WriteHeader(output.Code)
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Println("err encode: " + err.Error())
		}
	}
}

func createFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

		reader, err := r.MultipartReader()
		if err != nil {
			w.WriteHeader(statusNotFound)
			sendJSONErr(w, statusNotFound, err.Error())
			return
		}

		var request model.ShortenFileRequest

		for i := 0; i < 2; i++ {
			part, err := reader.NextPart()
			if err != nil {
				w.WriteHeader(statusNotFound)
				sendJSONErr(w, statusNotFound, err.Error())
				return
			}

			switch part.FormName() {
			case "json_field":
				err := decodeJSONLinkRequest(r, &request)
				if err != nil {
					w.WriteHeader(statusBadReq)
					sendJSONErr(w, statusBadReq, err.Error())
					return
				}
			case "file_field":

			}
		}
	}

}

func find(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param := mux.Vars(r)
		if _, ok := param["alias"]; !ok {
			w.WriteHeader(statusNotFound)
			sendJSONErr(w, statusNotFound, "not found")
			return
		}

		output := svc.Find(r.Context(), param["alias"])
		w.WriteHeader(output.Code)
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Println("err encode: " + err.Error())
		}
	}
}

func downloadFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func sendJSONErr(w io.Writer, code int, msg string) {
	m := map[string]any{
		"status": code,
		"msg":    msg,
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Println("err encode: " + err.Error())
	}
}

func decodeJSONLinkRequest[inType Input](r *http.Request, in *inType) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	defer r.Body.Close()

	if err := decoder.Decode(&in); err != nil {
		return err
	}

	if err := helper.ValidateStruct(*in); err != nil {
		return err
	}

	return nil
}

func checkLimitFileSize(part multipart.File) error {
	buf := bufio.NewReader(part)

	file, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	limit := io.MultiReader(buf, io.LimitReader(part, (10<<20)-511))
	written, err := io.Copy(file, limit)

	if err != nil && written > (10<<20) {
		return fmt.Errorf("max size reached: %w", err)
	}

	if err != nil {
		return err
	}

	return nil
}

type Input interface {
	model.ShortenRequest | model.ShortenFileRequest
}
