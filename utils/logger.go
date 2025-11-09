package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	LevelError LogLevel = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

func parseLevel(s string) LogLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

type Logger struct {
	level LogLevel
	l     *log.Logger
}

func NewLogger(level string) *Logger {
	return &Logger{
		level: parseLevel(level),
		l:     log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) logf(min LogLevel, prefix string, format string, args ...any) {
	if l.level < min {
		return
	}
	l.l.Printf("%s %s", prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) { l.logf(LevelDebug, "[DEBUG]", format, args...) }
func (l *Logger) Infof(format string, args ...any)  { l.logf(LevelInfo, "[INFO ]", format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.logf(LevelWarn, "[WARN ]", format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.logf(LevelError, "[ERROR]", format, args...) }

func (l *Logger) Debug(msg string) { l.Debugf("%s", msg) }
func (l *Logger) Info(msg string)  { l.Infof("%s", msg) }
func (l *Logger) Warn(msg string)  { l.Warnf("%s", msg) }
func (l *Logger) Error(msg string) { l.Errorf("%s", msg) }
