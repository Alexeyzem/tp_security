package handler

import (
	"fmt"
	"net/http"
	"strings"
)

func getHost(r *http.Request, port string) (string, error) {
	targetHost := r.URL.Host
	if targetHost == "" {
		targetHost = r.Host
	}
	if targetHost == "" {
		return "", fmt.Errorf("missing target host")
	}

	if !strings.Contains(targetHost, ":") {
		targetHost += ":" + port
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
