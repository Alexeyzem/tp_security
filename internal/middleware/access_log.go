package middleware

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	hijacker http.Hijacker
	status   int
	size     int
}

func (w *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.hijacker == nil {
		return nil, nil, http.ErrNotSupported
	}
	return w.hijacker.Hijack()
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func AccessLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}

			if hj, ok := w.(http.Hijacker); ok {
				lrw.hijacker = hj
			}

			handler.ServeHTTP(lrw, r)
			duration := time.Since(start)
			log.Printf(
				"%s %s %s | Headers: %v | Status: %d | Size: %d bytes | Duration: %s",
				r.Method,
				r.Host,
				r.URL.RequestURI(),
				r.Header,
				lrw.status,
				lrw.size,
				duration,
			)
		},
	)
}
