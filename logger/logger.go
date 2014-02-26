package logger

import (
	"fmt"
	"log"
	"os"
)

type Level uint8

const (
	Debug = iota
	Info
	Warn
	Error
	Fatal
	Panic
)

type logger struct {
	l   *log.Logger
	lvl Level
}

func New(prefix, filename string, lvl Level) *logger {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if os.IsNotExist(err) {
	// 	fd, err = os.Create(filename)
	// }
	if err != nil {
		panic(fmt.Sprintf("creating log file '%s', %v", filename, err))
	}
	return &logger{
		l:   log.New(fd, prefix, log.Lmicroseconds),
		lvl: lvl,
	}
}

func (l *logger) Debugf(msg string, arg ...interface{}) {
	if l.lvl > Debug {
		return
	}
	l.l.Printf("[Debug] "+msg, arg...)
}

func (l *logger) Infof(msg string, arg ...interface{}) {
	if l.lvl > Info {
		return
	}
	l.l.Printf("[Info] "+msg, arg...)
}
func (l *logger) Warnf(msg string, arg ...interface{}) {
	if l.lvl > Warn {
		return
	}
	l.l.Printf("[Warn] "+msg, arg...)
}
func (l *logger) Errorf(msg string, arg ...interface{}) {
	if l.lvl > Error {
		return
	}
	l.l.Printf("[Error] "+msg, arg...)
}

func (l *logger) Fatalf(msg string, arg ...interface{}) {
	if l.lvl > Fatal {
		return
	}
	l.l.Fatalf("[Fatal] "+msg, arg...)
}

func (l *logger) Panicf(msg string, arg ...interface{}) {
	if l.lvl > Panic {
		return
	}
	l.l.Panicf("[Panic] "+msg, arg...)
}
