package handler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	http1 = "HTTP/1.1\r\n"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Proxy-Connection")

	targetHost, err := getHost(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	targetConn, err := net.DialTimeout("tcp", targetHost, 5*time.Second)
	if err != nil {
		http.Error(w, "failed to connect to target", http.StatusBadGateway)

		return
	}
	defer targetConn.Close()

	_, err = fmt.Fprintf(targetConn, "%s %s %s", r.Method, getPath(r), http1)
	if err != nil {
		http.Error(w, "failed to write request", http.StatusBadGateway)
		return
	}

	r.Header.Set("Host", strings.Split(targetHost, ":")[0])
	err = r.Header.Write(targetConn)
	if err != nil {
		http.Error(w, "failed to write headers", http.StatusBadGateway)

		return
	}

	_, err = targetConn.Write([]byte("\r\n"))
	if err != nil {
		http.Error(w, "failed to write headers terminator", http.StatusBadGateway)

		return
	}

	if r.Body != nil {
		_, err = io.Copy(targetConn, r.Body)
		if err != nil {
			http.Error(w, "failed to write body", http.StatusBadGateway)

			return
		}
	}

	resp, err := http.ReadResponse(bufio.NewReader(targetConn), r)
	if err != nil {
		http.Error(w, "failed to read response", http.StatusBadGateway)

		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("error copying response body:", err)
	}
}

func getHost(r *http.Request) (string, error) {
	targetHost := r.URL.Host
	if targetHost == "" {
		targetHost = r.Host
	}
	if targetHost == "" {
		return "", fmt.Errorf("missing target host")
	}

	if !strings.Contains(targetHost, ":") {
		targetHost += ":80" //default for http
	}

	return targetHost, nil
}

func getPath(r *http.Request) string {
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}

	return path
}
