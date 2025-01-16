package logger

import (
	"log"
	"os"
)


var enabled bool

func init() {
	enabled = os.Getenv("LOG_ENABLED") == "true"
}

func Log(str string, v ...interface{}) {
	if enabled {
		log.Printf(str, v...)
	}
}