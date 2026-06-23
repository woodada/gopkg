package mysql

import (
	"context"
)

type Logger interface {
	CtxDebugf(ctx context.Context, format string, v ...interface{})
	CtxInfof(ctx context.Context, format string, v ...interface{})
	CtxWarnf(ctx context.Context, format string, v ...interface{})
	CtxErrorf(ctx context.Context, format string, v ...interface{})
}

var globalLogger Logger = defaultLogger{}

func SetLogger(ll Logger) {
	globalLogger = ll
}

type defaultLogger struct {
}

func (d defaultLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {

}

func (d defaultLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {

}

func (d defaultLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {

}

func (d defaultLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {

}
