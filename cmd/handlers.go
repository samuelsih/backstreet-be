package main

import (
	"backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"backstreetlinkv2/cmd/service"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

const (
	statusBadReq      = http.StatusBadRequest
	statusNotFound    = http.StatusNotFound
	statusInternalErr = http.StatusInternalServerError
	fileMaxSize       = 10 << 20
)

func createLink(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request model.ShortenRequest

		err := decodeJSONLinkRequest(r, &request)
		if err != nil {
			w.WriteHeader(statusBadReq)
			sendJSONErr(w, statusBadReq, err.Error())
			return
		}

		output := svc.InsertLink(r.Context(), request)
		w.WriteHeader(output.Code)
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Err(err)
		}
	}
}

func createFile(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		reader, err := r.MultipartReader()
		if err != nil {
			w.WriteHeader(statusBadReq)
			sendJSONErr(w, statusBadReq, err.Error())
			return
		}

		var request model.ShortenFileRequest

		for i := 0; i < 2; i++ {
			part, err := reader.NextPart()
			if err != nil {
				if closeErr := part.Close(); closeErr != nil {
					log.Err(closeErr)
				}

				w.WriteHeader(statusBadReq)
				sendJSONErr(w, statusBadReq, err.Error())
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
				if err := checkLimitFileSize(part); err != nil {
					w.WriteHeader(statusBadReq)
					sendJSONErr(w, statusBadReq, err.Error())
					return
				}

				request.RawFile = part
				request.Filename = part.FileName()

			default:
				err := fmt.Sprintf("unexpected field %v", part.FormName())
				w.WriteHeader(statusBadReq)
				sendJSONErr(w, statusBadReq, err)
				return
			}
		}

		output := svc.InsertFile(r.Context(), request)
		w.WriteHeader(output.Code)
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Err(err)
		}
	}

}

func find(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		param := mux.Vars(r)
		if _, ok := param["alias"]; !ok {
			w.WriteHeader(statusNotFound)
			sendJSONErr(w, statusNotFound, "not found")
			return
		}

		output := svc.Find(r.Context(), param["alias"])
		w.WriteHeader(output.Code)
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Err(err)
		}
	}
}

func downloadFile(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param := mux.Vars(r)
		if _, ok := param["alias"]; !ok {
			w.WriteHeader(statusNotFound)
			w.Header().Set("Content-Type", "application/json")
			sendJSONErr(w, statusNotFound, "not found")
			return
		}

		output := svc.DownloadFile(r.Context(), param["alias"])

		w.Header().Set("Content-Disposition", output.ContentDisposition)
		w.Header().Set("Content-Type", output.ContentType)
		w.Header().Set("Content-Length", output.ContentLength)

		_, err := io.Copy(w, output.File)
		if err != nil {
			log.Err(err)
			w.WriteHeader(statusInternalErr)
			sendJSONErr(w, statusInternalErr, err.Error())
		}
	}
}

func sendJSONErr(w io.Writer, code int, msg string) {
	m := map[string]any{
		"status": code,
		"msg":    msg,
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Err(err)
	}
}

func decodeJSONLinkRequest[inType helper.Request](r *http.Request, in *inType) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Err(err)
		}
	}()

	if err := decoder.Decode(&in); err != nil {
		return err
	}

	if err := helper.ValidateStruct(*in); err != nil {
		return err
	}

	return nil
}

func checkLimitFileSize(part io.Reader) error {
	buf := bufio.NewReader(part)

	file, err := os.Open("")
	if err != nil {
		return err
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Err(err)
		}
	}()

	limit := io.MultiReader(buf, io.LimitReader(part, fileMaxSize))
	written, err := io.Copy(file, limit)

	if err != nil && written > (10<<20) {
		return fmt.Errorf("max size reached: %w", err)
	}

	if err != nil {
		return err
	}

	return nil
}
