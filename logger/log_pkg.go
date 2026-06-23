package logger

import (
	"context"
)

var defLog *LogContainer = &LogContainer{
	textLog: Default(),
	jsonLog: nil,
}

func SetLevel(lv Level) {
	defLog.SetLevel(lv)
}

func DefaultLogger() *LogContainer {
	return defLog
}

func SetLogger(v *LogContainer) {
	defLog = v
}

func Clone() *LogContainer {
	return New(defLog.Opt)
}

// Error calls the default logger's Error method.
func Error(v ...interface{}) {
	defLog.Error(v...)
}
func Fatal(v ...any) {
	defLog.Fatal(v...)
}

// Warn calls the default logger's Warn method.
func Warn(v ...interface{}) {
	defLog.Warn(v...)
}

// Info calls the default logger's Info method.
func Info(v ...interface{}) {
	defLog.Info(v...)
}

// Debug calls the default logger's Debug method.
func Debug(v ...interface{}) {
	defLog.Debug(v...)
}
func Trace(v ...any) {
	defLog.Trace(v...)
}
func Notice(v ...any) {
	defLog.Notice(v...)
}

// Errorf calls the default logger's Errorf method.
func Errorf(format string, v ...interface{}) {
	defLog.Errorf(format, v...)
}
func Fatalf(format string, v ...any) {
	defLog.Fatalf(format, v...)
}

// Warnf calls the default logger's Warnf method.
func Warnf(format string, v ...interface{}) {
	defLog.Warnf(format, v...)
}

// Infof calls the default logger's Infof method.
func Infof(format string, v ...interface{}) {
	defLog.Infof(format, v...)
}

// Debugf calls the default logger's Debugf method.
func Debugf(format string, v ...interface{}) {
	defLog.Debugf(format, v...)
}
func Tracef(format string, v ...any) {
	defLog.Tracef(format, v...)
}
func Noticef(format string, v ...any) {
	defLog.Noticef(format, v...)
}

// CtxErrorf calls the default logger's CtxErrorf method.
func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	defLog.CtxErrorf(ctx, format, v...)
}
func CtxFatalf(ctx context.Context, format string, v ...any) {
	defLog.CtxFatalf(ctx, format, v...)
}

// CtxWarnf calls the default logger's CtxWarnf method.
func CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	defLog.CtxWarnf(ctx, format, v...)
}

// CtxInfof calls the default logger's CtxInfof method.
func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	defLog.CtxInfof(ctx, format, v...)
}

// CtxDebugf calls the default logger's CtxDebugf method.
func CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	defLog.CtxDebugf(ctx, format, v...)
}
func CtxTracef(ctx context.Context, format string, v ...any) {
	defLog.CtxTracef(ctx, format, v...)
}
func CtxNoticef(ctx context.Context, format string, v ...any) {
	defLog.CtxNoticef(ctx, format, v...)
}
