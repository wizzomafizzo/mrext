package service

import (
	"fmt"
	"log"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

type Logger struct {
	log         *log.Logger
	EnableDebug bool
}

func NewLogger(name string) *Logger {
	logFile := fmt.Sprintf(config.LogFileTemplate, name)

	return &Logger{
		log: log.New(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    1,
			MaxBackups: 2,
		}, "", log.LstdFlags),
	}
}

func (l *Logger) Info(format string, v ...any) {
	l.log.Println("INFO", fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(format string, v ...any) {
	l.log.Println("WARN", fmt.Sprintf(format, v...))
}

func (l *Logger) Error(format string, v ...any) {
	l.log.Println("ERROR", fmt.Sprintf(format, v...))
}

func (l *Logger) Debug(format string, v ...any) {
	if !l.EnableDebug {
		return
	}
	l.log.Println("DEBUG", fmt.Sprintf(format, v...))
}
