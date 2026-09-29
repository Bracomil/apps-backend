package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

type LogLevel int

const (
	DEBUG = iota + 1
	PRODUCTION
)

type Logger struct {
	mode int

	location *time.Location
	info     *log.Logger
	warn     *log.Logger
	error    *log.Logger
	debug    *log.Logger
	debug2   *log.Logger
}

func New(mode int) *Logger {
	flags := 0

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {

	}

	return &Logger{
		mode:     mode,
		location: location,
		info:     log.New(os.Stdout, "", flags),
		warn:     log.New(os.Stdout, "", flags),
		error:    log.New(os.Stderr, "", flags),
		debug:    log.New(os.Stdout, "", flags),
		debug2:   log.New(os.Stdout, "", flags),
	}
}

func (l *Logger) Info(v ...interface{}) {
	v = append([]interface{}{fmt.Sprintf("%s - [INFO] ", (time.Now().In(l.location)).Format("02-01-2006 15:04:05"))}, v...)
	l.info.Println(v...)
}

func (l *Logger) Warn(v ...interface{}) {
	// _, file, line, _ := runtime.Caller(1)
	// v = append([]interface{}{fmt.Sprintf("%s:%d ", file, line)}, v...)
	v = append([]interface{}{fmt.Sprintf("%s - [WARN] ", (time.Now().In(l.location)).Format("02-01-2006 15:04:05"))}, v...)

	l.warn.Println(v...)
}

func (l *Logger) Error(v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	v = append([]interface{}{fmt.Sprintf("%s:%d ", file, line)}, v...)
	v = append([]interface{}{fmt.Sprintf("%s - [ERROR] ", (time.Now().In(l.location)).Format("02-01-2006 15:04:05"))}, v...)

	l.error.Println(v...)
}

func (l *Logger) Debug(v ...interface{}) {
	if l.mode == DEBUG {
		v = append([]interface{}{fmt.Sprintf("%s - [DEBUG] ", (time.Now().In(l.location)).Format("02-01-2006 15:04:05"))}, v...)
		l.debug.Println(v...)
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	v = append([]interface{}{fmt.Sprintf("%s:%d ", file, line)}, v...)
	v = append([]interface{}{fmt.Sprintf("%s - [FATAL] ", (time.Now().In(l.location)).Format("02-01-2006 15:04:05"))}, v...)
	l.error.Fatal(v...)
}
