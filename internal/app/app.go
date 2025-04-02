package app

import (
	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/handler"
	"github.com/tp_security/internal/middleware"
	"log"
	"net/http"
)

func Run(cfg *config.Config) error {
	server := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.AccessLog(handler.New()),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("Starting HTTP proxy server on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
