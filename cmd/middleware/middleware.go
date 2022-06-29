package middleware

import (
	"encoding/json"
	"net/http"
)

func Recoverer(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// susah handle yang satu ini, jadi dipanic aja
					panic(rvr)
				}

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func SetResponseHeaderJSON(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func OnlyJSONRequest(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Content-Type")

		if header != "application/json" {
			json.NewEncoder(w).Encode(map[string]string{"error": "header must be application/json"})
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}