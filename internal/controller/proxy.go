package controller

import (
	"io"
	"net/http"

	"github.com/tp_security/internal/config"
)

type HttpSender interface {
	Do(req *http.Request) (*http.Response, error)
}

type Proxy struct {
	client HttpSender
}

func NewProxy(cfg *config.Config) *Proxy {
	return &Proxy{
		client: &http.Client{
			Timeout:   cfg.WriteTimeout,
			Transport: http.DefaultTransport,
		},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Proxy-Connection")
	r.RequestURI = ""
	resp, err := p.client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
