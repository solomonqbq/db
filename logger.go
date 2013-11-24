package db

import (
	"fmt"
	"log"
	"os"
)

type logger struct {
	log *log.Logger
}

func standardLogger() *logger {
	return &logger{
		log: log.New(os.Stderr, "[DB] ", log.Ltime|log.Lshortfile),
	}
}

func (l *logger) Debug(format string, args ...interface{}) {
	format = fmt.Sprintf("\033[1;30m%s\033[0m", format)
	l.log.Output(3, fmt.Sprintf(format, args...))
}

func (l *logger) Info(format string, args ...interface{}) {
	l.log.Output(3, fmt.Sprintf(format, args...))
}

func (l *logger) Warn(format string, args ...interface{}) {
	format = fmt.Sprintf("\033[33m%s\033[0m", format)
	l.log.Output(3, fmt.Sprintf(format, args...))
}

func (l *logger) Error(format string, args ...interface{}) {
	format = fmt.Sprintf("\033[31m%s\033[0m", format)
	l.log.Output(3, fmt.Sprintf(format, args...))
}
