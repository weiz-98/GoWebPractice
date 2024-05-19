package main

import (
	"net/http"
	"time"
)

func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

type warpWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *warpWriter) writeHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &warpWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		app.infoLog.Printf("%d %s %s %s", wrapped.statusCode, r.Method, r.URL.RequestURI(), time.Since(start))
		next.ServeHTTP(wrapped, r)
	})
}
