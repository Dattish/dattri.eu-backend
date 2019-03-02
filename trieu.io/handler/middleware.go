package handler

import (
	"log"
	"net/http"
	"time"
)

func CORS(allowedMethods string, allowedOrigin string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		handler.ServeHTTP(w, r)
	})
}

func CSP(policy string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", policy)
		handler.ServeHTTP(w, r)
	})
}

func Logging(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer log.Printf("[%s] %s: %s | %v", r.RemoteAddr, r.Method, r.URL, time.Since(start))
		handler.ServeHTTP(w, r)
	})
}
