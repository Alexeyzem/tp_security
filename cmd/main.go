package main

import (
	"github.com/tp_security/internal/app"
	"log"

	"github.com/tp_security/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
