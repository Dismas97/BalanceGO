package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	bytesSent  int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesSent += n
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error leyendo body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		
		wrapped := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)

		log.Printf("[%s] %s %s - Estado: %d, Tamaño: %d bytes, Duracion: %v, IP: %s, Peticion:\n%v",
			r.Method,
			r.URL.Path,
			r.Proto,
			wrapped.statusCode,
			wrapped.bytesSent,
			duration,
			r.RemoteAddr,
			string(body),
		)
	})
}
