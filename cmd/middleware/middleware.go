package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(time.Second), 3)

func Recoverer(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				fmt.Println("PANIC\t", rvr)
				
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

func Limit(next http.Handler) http.Handler {
	f := func (w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "too many requests, try again later"})
			return
		}

		next.ServeHTTP(w, r)

	}

	return http.HandlerFunc(f)
}