package main

import (
	"log"

	"github.com/kevwargo/bootsnap/internal/bootsnap"
)

func main() {
	if err := bootsnap.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
