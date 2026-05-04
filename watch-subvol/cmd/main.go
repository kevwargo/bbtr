package main

import (
	"log"
	"os"

	"github.com/kevwargo/btrscr/watch-subvol/internal/watcher"
)

func main() {
	if err := watcher.Watch(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
