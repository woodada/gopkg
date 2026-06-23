package logger

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"io"
	"os"
	"runtime"
	"strings"
)

type nLog struct {
	slog         *slog.Logger
	slogLv       *slog.LevelVar
	ctxHandle    CtxHandle
	AddSource    bool
	SourceOffset int
	output       io.Writer
}

func (nl nLog) getCtxVal(ctx context.Context) []Attr {
	if nl.ctxHandle == nil {
		return nil
	}
	ctxVal := nl.ctxHandle.GetCtxVal(ctx)
	ctxV := make([]slog.Attr, 0)
	for _, v := range ctxVal {
		ctxV = append(ctxV, v)
	}
	return ctxV
}

func (nl nLog) getLineAttr() slog.Attr {
	if !nl.AddSource {
		return slog.Attr{}
	}
	offset := 7 + nl.SourceOffset
	_, file, line, _ := runtime.Caller(offset)
	file = SpiltFilePath(file)
	return slog.String(LineKey, fmt.Sprintf("%s:%d", file, line))
}

func (nl nLog) getSlogOpt(ctx context.Context, v []any) (msg string, opt []any) {
	msg = ""
	n := make([]any, 0)
	hasFile := false
	if v != nil {
		for _, c := range v {
			switch val := c.(type) {
			case slog.Attr:
				if val.Key == LineKey {
					hasFile = true
					c = slog.String(val.Key, SpiltFilePath(val.Value.String()))
				}
				n = append(n, c)
			default:
				msg += " " + fmt.Sprint(c)
			}
		}
	}
	if nl.AddSource && !hasFile {
		n = append(n, nl.getLineAttr())
	}
	ctxVal := nl.getCtxVal(ctx)
	if len(ctxVal) > 0 {
		for _, cv := range ctxVal {
			n = append(n, cv)
		}
	}
	return strings.TrimLeft(msg, " "), n
}

func (nl nLog) splitAttrFormat(v []any) (formatArr []any, attrArr []any) {
	formatArr = make([]any, 0)
	attrArr = make([]any, 0)
	if v != nil {
		for _, c := range v {
			if _, ok := c.(slog.Attr); ok {
				attrArr = append(attrArr, c)
			} else {
				formatArr = append(formatArr, c)
			}
		}
	}
	return formatArr, attrArr
}

func Default() FullLogger {
	logLv := &slog.LevelVar{}
	iow := os.Stdout
	return nLog{
		slog:      newSLog(iow, logLv, LogTypeText, nil),
		slogLv:    logLv,
		ctxHandle: DefCtxHandle,
		output:    iow,
	}
}

func (nl nLog) Debug(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	nl.slog.Debug(msg, opt...)
}
func (nl nLog) Info(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	nl.slog.Info(msg, opt...)
}
func (nl nLog) Warn(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	nl.slog.Warn(msg, opt...)
}
func (nl nLog) Error(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	nl.slog.Error(msg, opt...)
}
func (nl nLog) Fatal(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	nl.slog.Error(msg, opt...)
}
func (nl nLog) Trace(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	opt = append(opt, slog.String(SecondLevel, "Trace"))
	nl.slog.Debug(msg, opt...)
}
func (nl nLog) Notice(v ...any) {
	msg, opt := nl.getSlogOpt(context.Background(), v)
	opt = append(opt, slog.String(SecondLevel, "Notice"))
	nl.slog.Warn(msg, opt...)
}

func (nl nLog) Debugf(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Debug(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) Infof(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Info(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) Warnf(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Warn(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) Errorf(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Error(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) Fatalf(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Error(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) Tracef(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	attrArr = append(attrArr, slog.String(SecondLevel, "Trace"))
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Error(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) Noticef(format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	attrArr = append(attrArr, slog.String(SecondLevel, "Notice"))
	_, opt := nl.getSlogOpt(context.Background(), attrArr)
	nl.slog.Warn(fmt.Sprintf(format, formatArr...), opt...)
}

func (nl nLog) CtxDebugf(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Debug(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) CtxInfof(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Info(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) CtxWarnf(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Warn(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) CtxErrorf(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Error(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) CtxFatalf(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Error(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) CtxTracef(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	attrArr = append(attrArr, slog.String(SecondLevel, "Trace"))
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Error(fmt.Sprintf(format, formatArr...), opt...)
}
func (nl nLog) CtxNoticef(ctx context.Context, format string, v ...any) {
	formatArr, attrArr := nl.splitAttrFormat(v)
	attrArr = append(attrArr, slog.String(SecondLevel, "Notice"))
	_, opt := nl.getSlogOpt(ctx, attrArr)
	nl.slog.Warn(fmt.Sprintf(format, formatArr...), opt...)
}

func (nl nLog) SetLevel(lv Level) {
	nl.slogLv.Set(lv)
}

func (nl nLog) SetOutput(iow io.Writer) {
	nl.output = iow
}
