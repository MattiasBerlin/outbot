package main

import (
	"fmt"
	"io"
	"time"
)

// Logger writes logs on different levels with timestamps.
type Logger struct {
	output io.Writer
}

// New logger.
func New(output io.Writer) Logger {
	return Logger{
		output: output,
	}
}

func (l Logger) write(level string, f string, args ...interface{}) string {
	formatted := fmt.Sprintf(f, args)
	text := fmt.Sprintf("%v %v %v", time.Now().Format(time.RFC3339), level, formatted)
	_, err := l.output.Write([]byte(text))
	if err != nil {
		panic("Unable to write to log file")
	}

	return formatted
}

// Info writes a log on info level.
func (l Logger) Info(f string, args ...interface{}) {
	l.write("INFO", f, args...)
}

// Warn writes a log on warning level.
func (l Logger) Warn(f string, args ...interface{}) {
	l.write("WARNING", f, args...)
}

// Error writes a log on error level.
func (l Logger) Error(f string, args ...interface{}) {
	l.write("ERROR", f, args...)
}

// Fatal writes a log on fatal level.
// After logging a panic will occur.
func (l Logger) Fatal(f string, args ...interface{}) {
	text := l.write("FATAL", f, args...)
	panic(text)
}
