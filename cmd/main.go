package main

import (
	"github.com/dontpanicw/EventBooker/config"
	"github.com/dontpanicw/EventBooker/internal/app"
	"log"
)

func main() {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("error creating config")
	}

	if err := app.Start(cfg); err != nil {
		log.Fatal("failed to start application")
	}

}
