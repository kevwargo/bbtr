package main

import (
	"log"

	"github.com/kevwargo/btrscr/bootsnap/internal/bootsnap"
	"github.com/kevwargo/btrscr/bootsnap/internal/logger"
)

func main() {
	logger.Init()

	if err := bootsnap.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
