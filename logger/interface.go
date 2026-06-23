package logger

import (
	"context"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.opentelemetry.io/otel/trace"
	"io"
	"os"
	"time"
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	Trace(v ...interface{})
	Notice(v ...interface{})
}

// FormatLogger is a logger interface that output logs with a format.
type FormatLogger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Tracef(format string, v ...interface{})
	Noticef(format string, v ...interface{})
}

type CtxLogger interface {
	CtxDebugf(ctx context.Context, format string, v ...interface{})
	CtxInfof(ctx context.Context, format string, v ...interface{})
	CtxWarnf(ctx context.Context, format string, v ...interface{})
	CtxErrorf(ctx context.Context, format string, v ...interface{})
	CtxFatalf(ctx context.Context, format string, v ...interface{})
	CtxTracef(ctx context.Context, format string, v ...interface{})
	CtxNoticef(ctx context.Context, format string, v ...interface{})
}

// Control provides methods to config a logger.
type Control interface {
	SetLevel(Level)
	SetOutput(io.Writer)
}

// FullLogger is the combination of Logger, FormatLogger, CtxLogger and Control.
type FullLogger interface {
	Logger
	FormatLogger
	CtxLogger
	Control
}

type CtxHandle interface {
	GetCtxVal(ctx context.Context) []Attr
}

var DefCtxHandle CtxHandle = defCtxHandle{}

type defCtxHandle struct {
}

func (dc defCtxHandle) GetCtxVal(ctx context.Context) []Attr {
	ctxArr := make([]Attr, 0)
	span := trace.SpanFromContext(ctx)
	//if span.SpanContext().IsValid() {
	if span != nil {
		traceId := span.SpanContext().TraceID().String()
		if traceId != "" {
			ctxArr = append(ctxArr, String(CtxTraceId, traceId))
		}
		spanId := span.SpanContext().SpanID().String()
		if spanId != "" {
			ctxArr = append(ctxArr, String(CtxSpanId, spanId))
		}
	}
	//if ctx != nil {
	//	if traceId, ok := ctx.Value(CtxTraceId).(string); ok {
	//		ctxArr = append(ctxArr, String(CtxTraceId, traceId))
	//	}
	//	if spanId, ok := ctx.Value(CtxSpanId).(string); ok {
	//		ctxArr = append(ctxArr, String(CtxSpanId, spanId))
	//	}
	//}
	return ctxArr
}

type LogWrite struct {
	callbackFunc func(p []byte) (n int, err error)
}

func (lw *LogWrite) Write(p []byte) (n int, err error) {
	if lw.callbackFunc != nil {
		return lw.callbackFunc(p)
	}
	return 0, nil
}

func NewLogWrite(f func(p []byte) (n int, err error)) *LogWrite {
	return &LogWrite{
		callbackFunc: f,
	}
}

type OutInputLog interface {
	GetLogWriter() io.Writer
}

type lOutInput struct {
	RotationTime time.Duration // 日志轮转时间
	RotationSize LogSize       // 日志大小轮转
	LogPath      string
	writer       io.Writer
}

func (pl *lOutInput) initWriter() {
	if pl.writer != nil {
		return
	}
	if pl.LogPath == "" && pl.RotationSize == 0 {
		pl.writer = os.Stdout
		return
	}
	rotateOpt := []rotatelogs.Option{
		rotatelogs.WithLinkName(pl.LogPath),
		//rotatelogs.WithMaxAge(time.Duration(180)*time.Second),      // 保留最近 3
	}
	if pl.RotationTime != 0 {
		rotateOpt = append(rotateOpt, rotatelogs.WithRotationTime(pl.RotationTime)) // 时间轮转一个新文件
	}
	if pl.RotationSize != 0 {
		rotateOpt = append(rotateOpt, rotatelogs.WithRotationSize(pl.RotationSize)) // 文件达到多大则进行切割，单位为 bytes
	}
	format := "%Y%m%d%H%M%S"
	if pl.RotationTime >= time.Minute {
		format = "%Y%m%d%H%M"
	}
	if pl.RotationTime >= time.Hour {
		format = "%Y%m%d%H"
	}
	if pl.RotationTime >= 24*time.Hour {
		format = "%Y%m%d"
	}
	if pl.RotationTime >= 30*24*time.Hour {
		format = "%Y%m"
	}
	if pl.RotationTime >= 365*24*time.Hour {
		format = "%Y"
	}
	writer, _ := rotatelogs.New(
		pl.LogPath+"."+format,
		rotateOpt...,
	)
	pl.writer = writer
}

func (pl *lOutInput) GetLogWriter() io.Writer {
	pl.initWriter()
	return pl.writer
}

var DefLogOutInput OutInputLog = &lOutInput{}
