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

func Stream() io.Writer {
	if stream != nil {
		return stream
	}

	stream = os.Stderr

	f, err := os.OpenFile("/run/initramfs/bootsnap.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err == nil {
		stream = io.MultiWriter(stream, f)
	}

	return stream
}

var (
	stream io.Writer
	logger *log.Logger
)

func getLogger() *log.Logger {
	if logger == nil {
		logger = log.New(Stream(), "", log.LstdFlags)
	}

	return logger
}
