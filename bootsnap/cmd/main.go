package main

import (
	"log"

	"github.com/kevwargo/bbtr/bootsnap/internal/bootsnap"
	"github.com/kevwargo/bbtr/bootsnap/internal/logger"
)

func main() {
	logger.Init()

	if err := bootsnap.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
