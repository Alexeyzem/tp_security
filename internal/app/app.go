package app

import (
	"crypto/tls"
	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/handler"
	"github.com/tp_security/internal/middleware"
	"log"
	"net/http"
)

func Run(cfg *config.Config) error {
	handleFunc, err := handler.New(cfg)
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.AccessLog(handleFunc),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("Starting HTTP/HTTPS proxy server on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
