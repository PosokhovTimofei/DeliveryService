package utils

import "net/http"

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter, code int) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, code}
}

func (lwr *LoggingResponseWriter) WriteHeader(code int) {
	lwr.StatusCode = code
	lwr.ResponseWriter.WriteHeader(code)
}
