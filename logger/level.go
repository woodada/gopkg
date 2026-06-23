package logger

import (
	"golang.org/x/exp/slog"
	"strings"
)

type Leveler = slog.Leveler

type Level = slog.Level

const (
	LevelDebug Level = slog.LevelDebug
	LevelInfo  Level = slog.LevelInfo
	LevelWarn  Level = slog.LevelWarn
	LevelError Level = slog.LevelError
)

func GetLevel(lvStr string) Level {
	switch strings.ToLower(lvStr) {
	case strings.ToLower(LevelDebug.String()):
		return slog.LevelDebug
	case strings.ToLower(LevelInfo.String()):
		return slog.LevelInfo
	case strings.ToLower(LevelWarn.String()):
		return slog.LevelWarn
	case strings.ToLower(LevelError.String()):
		return slog.LevelError
	}
	return slog.LevelInfo
}
