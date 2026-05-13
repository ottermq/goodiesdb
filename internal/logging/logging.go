package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type Level int

const (
	LevelError Level = iota
	LevelInfo
	LevelDebug
)

var (
	mu           sync.RWMutex
	currentLevel = LevelInfo
)

func SetLevel(raw string) error {
	level, err := parseLevel(raw)
	if err != nil {
		return err
	}

	mu.Lock()
	currentLevel = level
	mu.Unlock()
	return nil
}

func Debugf(format string, args ...any) {
	logf(LevelDebug, "DEBUG", format, args...)
}

func Infof(format string, args ...any) {
	logf(LevelInfo, "INFO", format, args...)
}

func Errorf(format string, args ...any) {
	logf(LevelError, "ERROR", format, args...)
}

func logf(level Level, label string, format string, args ...any) {
	mu.RLock()
	enabled := level <= currentLevel
	mu.RUnlock()
	if !enabled {
		return
	}

	fmt.Fprintf(os.Stdout, "[%s] %s\n", label, fmt.Sprintf(format, args...))
}

func parseLevel(raw string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("invalid log level %q", raw)
	}
}
