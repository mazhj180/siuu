package http

import ht "net/http"

// statusWriter wraps http.ResponseWriter to capture status code
type StatusWriter struct {
	ht.ResponseWriter
	StatusCode int
}

func (w *StatusWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}
