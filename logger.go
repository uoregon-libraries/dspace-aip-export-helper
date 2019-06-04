package main

import (
	"log"
	"os"
)

// A Logger includes anything with the ability to print in some way, including
// the standard log.Logger structure
type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
}

// nullLogger implement Logger but does nothing with the outputs
type nullLogger struct{}

func (nullLogger) Printf(string, ...interface{}) {}
func (nullLogger) Print(...interface{})          {}
func (nullLogger) Println(...interface{})        {}

// NullLogger may be used as a Logger that ignores all calls
var NullLogger nullLogger

// StderrLogger returns a new log that just prints to Stderr in the same format
// as the default log.Logger, but keeps global state out of the picture
func StderrLogger() *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}
