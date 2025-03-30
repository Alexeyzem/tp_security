package main

import (
	"log"

	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/proxy_server"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := proxy_server.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
