package main

import (
	"github.com/kevwargo/bootsnap/internal/bootsnap"
	"github.com/kevwargo/bootsnap/internal/log"
)

func main() {
	if err := bootsnap.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
