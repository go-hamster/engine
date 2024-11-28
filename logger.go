package app

import (
	"log"
)

type Logger interface {
	Log(format string, args ...any)
}

type defaultLogger struct{}

func (l *defaultLogger) Log(format string, args ...any) {
	log.Printf(format, args...)
}
