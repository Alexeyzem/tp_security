package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tp_security/internal/core"
)

func (h *Handler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Proxy-Connection")

	targetHost, err := getHost(r, httpPort)
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

	r.Host = targetHost
	r.URL.Host = targetHost
	r.URL.Path = getPath(r)
	saveReq := parseRequest(r)
	saveResp := parseResponse(resp)

	err = h.repo.Save(r.Context(), saveReq, saveResp)
	if err != nil {
		log.Println("failed to save request:", err)
	}
}

func parseRequest(r *http.Request) string {
	req := &core.Request{
		Method:    r.Method,
		Path:      r.URL.Path,
		Host:      r.URL.Host,
		Headers:   toMapStringSlice(r.Header),
		GetParams: toMapStringSlice(r.URL.Query()),
	}

	cookies := r.Cookies()
	for _, cookie := range cookies {
		req.Cookies[cookie.Name] = cookie.Value
	}

	err := r.ParseForm()
	if err != nil {
		log.Println("error parsing form:", err)
	} else {
		params := r.Form
		req.PostParams = toMapStringSlice(params)
	}

	res, err := json.Marshal(req)
	if err != nil {
		log.Println("error marshaling request:", err)
	}

	return string(res)
}

func parseResponse(r *http.Response) string {
	resp := &core.Response{
		Code:    r.StatusCode,
		Headers: toMapStringSlice(r.Header),
		Message: r.Status,
	}

	var body []byte
	_, err := r.Body.Read(body)
	if err != nil {
		log.Println("error reading response body:", err)
	}
	resp.Body = string(body)

	res, err := json.Marshal(resp)
	if err != nil {
		log.Println("error marshaling response:", err)
	}

	return string(res)
}

func toMapStringSlice[IN ~map[string][]string](in IN) map[string][]string {
	res := make(map[string][]string, len(in))
	for k, v := range in {
		for _, value := range v {
			res[k] = append(res[k], value)
		}
	}

	return res
}
