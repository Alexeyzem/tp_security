package proxy_server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/controller"
	"github.com/tp_security/internal/middleware"
)

func New(cfg *config.Config) (*http.Server, error) {
	tlsConfig := &tls.Config{}
	if cfg.UseTLS {
		tlsConfig = newTLSConfig(cfg)
	}
	cfg.TLSConfig = tlsConfig

	handler, err := controller.NewProxy(cfg)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      middleware.AccessLog(handler),
		TLSConfig:    tlsConfig,
	}, nil
}

func newTLSConfig(cfg *config.Config) *tls.Config {
	return &tls.Config{
		ServerName: cfg.Host,
		MinVersion: tls.VersionTLS13,
	}
}

func Run(cfg *config.Config) error {
	server, err := New(cfg)
	if err != nil {
		return err
	}

	log.Printf("Starting proxy server at port:%s", cfg.Port)
	if cfg.UseTLS {
		return server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}

	return server.ListenAndServe()
}
