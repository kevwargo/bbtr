package main

import (
	"log"
	"os"

	"github.com/kevwargo/bbtr/watch-subvol/internal/watcher"
)

func main() {
	switch len(os.Args) {
	case 2:
		if err := watcher.Watch(os.Args[1]); err != nil {
			log.Fatal(err)
		}
	case 3:
		if err := watcher.Diff(os.Args[1], os.Args[2]); err != nil {
			log.Fatal(err)
		}
	}
}
