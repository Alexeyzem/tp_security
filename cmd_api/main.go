package main

import (
	"log"

	"github.com/tp_security/internal/app"

	"github.com/tp_security/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.RunApi(cfg); err != nil {
		log.Fatal(err)
	}
}
