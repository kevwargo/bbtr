package logger

import (
	"io"
	"log"
	"os"
)

func Init() {
	var w io.Writer = os.Stderr

	f, err := os.OpenFile("/run/initramfs/bootsnap.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err == nil {
		w = io.MultiWriter(w, f)
	}

	log.SetOutput(w)
}
