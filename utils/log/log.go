package log

import (
	"io"
	logger "log"
	"os"
	"runtime"
	"strings"
	"github.com/rocketsoftware/open-web-launch/utils"
)

var newline string

func Fatal(err error) {
	Println(err)
	utils.ShowFatalError(err.Error())
	os.Exit(2)
}

func Println(v ...interface{}) {
	args := append(v, newline)
	logger.Println(args...)
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
