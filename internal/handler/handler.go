package handler

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/tp_security/internal/config"
)

const (
	http1          = "HTTP/1.1\r\n"
	successConnect = "HTTP/1.0 200 Connection established\r\n\r\n"
	httpPort       = "80"
	httpsPort      = "443"
)

type Handler struct {
	caCert     *x509.Certificate
	caKey      interface{}
	caCertPool *x509.CertPool
}

func New(cfg *config.Config) (*Handler, error) {
	handler := &Handler{}
	err := handler.loadCA(cfg.CertFile, cfg.KeyFile)

	return handler, err
}

func (h *Handler) loadCA(certFile, keyFile string) error {
	caCertPEM, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("reading CA cert: %w", err)
	}

	block, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("parsing CA cert: %w", err)
	}
	h.caCert = caCert

	caKeyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("reading CA key: %w", err)
	}

	block, _ = pem.Decode(caKeyPEM)
	caKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		caKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("parsing CA key: %w", err)
		}
	}
	h.caKey = caKey

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	h.caCertPool = caCertPool

	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		h.handleHTTPS(w, r)
		return
	}

	h.handleHTTP(w, r)
}
