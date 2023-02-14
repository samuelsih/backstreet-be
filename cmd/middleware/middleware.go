package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/rs/cors"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(time.Second), 3)

// Recoverer will recover the program when panic is triggered
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

// SetResponseHeaderJSON will set the header of the Response Writer to "Content-Type: application/json"
func SetResponseHeaderJSON(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

// OnlyJSONRequest will reject the request when the header request is not application/json
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

// Limit implements the rate limiter for the server
// The value of rate limiter is 3 request/s
func Limit(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "too many requests, try again later"})
			return
		}

		next.ServeHTTP(w, r)

	}

	return http.HandlerFunc(f)
}

func CORS(environment string) func(handler http.Handler) http.Handler {
	if environment == "PRODUCTION" {
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"https://backstreet.link", "https://www.backstreet.link"},
			AllowedMethods: []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Origin", "Authorization"},
			Debug:          false,
		})

		return func(handler http.Handler) http.Handler {
			return c.Handler(handler)
		}
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Authorization"},
		Debug:          false,
	})

	return func(handler http.Handler) http.Handler {
		return c.Handler(handler)
	}
}
