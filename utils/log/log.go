package log

import (
	"io"
	logger "log"
	"runtime"
	"strings"
)

var newline string

func Fatal(v ...interface{}) {
	logger.Fatal(v, newline)
}
func Println(v ...interface{}) {
	logger.Println(v, newline)
}

func Printf(format string, v ...interface{}) {
	if strings.HasSuffix(format, "\n") {
		format = strings.TrimSuffix(format, "\n")
	}
	logger.Printf(format+newline, v...)
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

func init() {
	if runtime.GOOS == "windows" {
		newline = "\r"
    }
    logger.SetFlags(logger.LstdFlags | logger.Lmicroseconds)
}
