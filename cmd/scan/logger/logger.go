package logger

import (
	"log"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
)

type Log struct{}

func (l *Log) Debug(message string, args ...interface{}) {
	if config.Debug {
		log.Printf(message+"\n", args...)
	}
}
