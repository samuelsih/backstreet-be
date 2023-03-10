package api

import (
	"backstreetlinkv2/api/helper"
	"backstreetlinkv2/api/model"
	"backstreetlinkv2/api/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
)

const (
	statusBadReq      = http.StatusBadRequest
	statusNotFound    = http.StatusNotFound
	statusInternalErr = http.StatusInternalServerError
	fileMaxSize       = 10 << 20
)

func CreateLink(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request model.ShortenRequest

		err := decodeJSONLinkRequest(r.Body, &request)
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

func CreateFile(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		r.Body = http.MaxBytesReader(w, r.Body, fileMaxSize)
		if err := r.ParseMultipartForm(fileMaxSize); err != nil {
			w.WriteHeader(statusBadReq)
			sendJSONErr(w, statusBadReq, err.Error())
			return
		}

		defer r.Body.Close()

		var request model.ShortenFileRequest

		jsonReader := strings.NewReader(r.FormValue("json_field"))
		err := decodeJSONLinkRequest(jsonReader, &request)
		if err != nil {
			w.WriteHeader(statusBadReq)
			sendJSONErr(w, statusBadReq, err.Error())
			return
		}

		file, header, err := r.FormFile("file_field")
		if err != nil {
			w.WriteHeader(statusBadReq)
			sendJSONErr(w, statusBadReq, err.Error())
			return
		}

		request.RawFile = file
		request.Filename = header.Filename

		output := svc.InsertFile(r.Context(), request)
		w.WriteHeader(output.Code)
		if err := json.NewEncoder(w).Encode(output); err != nil {
			log.Err(err)
		}
	}

}

func Find(svc *service.Deps) http.HandlerFunc {
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

func DownloadFile(svc *service.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param := mux.Vars(r)
		if _, ok := param["alias"]; !ok {
			w.WriteHeader(statusNotFound)
			w.Header().Set("Content-Type", "application/json")
			sendJSONErr(w, statusNotFound, "not found")
			return
		}

		output := svc.DownloadFile(r.Context(), param["alias"])
		if output.Code >= 400 && output.Code <= 599 {
			w.WriteHeader(output.Code)
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(output); err != nil {
				log.Err(err)
			}
			return
		}

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

func decodeJSONLinkRequest[inType helper.Request](r io.Reader, in *inType) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&in); err != nil {
		return err
	}

	if err := helper.ValidateStruct(*in); err != nil {
		return err
	}

	return nil
}
