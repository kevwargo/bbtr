package log

import (
	"io"
	"log"
	"os"
)

func Printf(format string, v ...any) {
	getLogger().Printf(format, v...)
}

func Println(v ...any) {
	getLogger().Println(v...)
}

func Fatal(v ...any) {
	getLogger().Fatal(v...)
}

var logger *log.Logger

func getLogger() *log.Logger {
	if logger != nil {
		return logger
	}

	var logW io.Writer = os.Stderr

	f, err := os.OpenFile("/run/initramfs/bootsnap.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err == nil {
		logW = io.MultiWriter(logW, f)
	}

	logger = log.New(logW, "", log.LstdFlags)

	return logger
}
