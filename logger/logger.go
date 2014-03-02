package logger

import (
	"fmt"
	"log"
	"os"
)

type Level uint8

const (
	Panic = iota
	Fatal
	Error
	Warn
	Info
	Debug
)

type Logger struct {
	l   *log.Logger
	lvl Level
}

func New(prefix, filename string, lvl Level) *Logger {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("creating log file '%s', %v", filename, err))
	}
	return &Logger{
		l:   log.New(fd, prefix, log.Lmicroseconds),
		lvl: lvl,
	}
}

func (l *Logger) Debugf(msg string, arg ...interface{}) {
	if l.lvl < Debug {
		return
	}
	l.l.Printf("[Debug] "+msg, arg...)
}

func (l *Logger) Infof(msg string, arg ...interface{}) {
	if l.lvl < Info {
		return
	}
	l.l.Printf("[Info] "+msg, arg...)
}
func (l *Logger) Warnf(msg string, arg ...interface{}) {
	if l.lvl < Warn {
		return
	}
	l.l.Printf("[Warn] "+msg, arg...)
}
func (l *Logger) Errorf(msg string, arg ...interface{}) {
	if l.lvl < Error {
		return
	}
	l.l.Printf("[Error] "+msg, arg...)
}

func (l *Logger) Fatalf(msg string, arg ...interface{}) {
	if l.lvl < Fatal {
		return
	}
	l.l.Fatalf("[Fatal] "+msg, arg...)
}

func (l *Logger) Panicf(msg string, arg ...interface{}) {
	l.l.Panicf("[Panic] "+msg, arg...)
}
