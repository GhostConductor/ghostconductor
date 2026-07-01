package main

import (
	"log"
	"net/http"
	"time"
)

// RequestLogger logs method, path, status, and duration for every request
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

// Recovery catches panics and returns 500 instead of crashing the process
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequireReadAccessForAPI validates API tokens from gc-api
// TODO: implement token validation against gc-api when token generation is ready
// For now, passthrough — all requests are allowed
func RequireReadAccessForAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Future: extract token from Authorization header
		// Future: validate token against gcapi.ValidateToken()
		next.ServeHTTP(w, r)
	}
}

// statusWriter wraps ResponseWriter to capture the status code for logging
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
