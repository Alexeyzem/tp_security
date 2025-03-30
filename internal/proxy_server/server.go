package proxy_server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/controller"
)

func New(cfg *config.Config) (*http.Server, error) {
	tlsConfig := &tls.Config{}
	if cfg.UseTLS {
		tlsConfig = newTLSConfig(cfg)
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      controller.NewProxy(cfg),
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
		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}
