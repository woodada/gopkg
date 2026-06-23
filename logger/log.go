package logger

import (
	"context"
	"io"
	"os"

	"golang.org/x/exp/slog"
)

type LogContainer struct {
	Opt     LogOptions
	jsonLog FullLogger
	textLog FullLogger
	LogLv   slog.Leveler
}

func (lb *LogContainer) wJson(b []byte) (n int, err error) {
	if lb.Opt.OutInput == nil {
		return 0, nil
	}
	for _, v := range lb.Opt.OutInput {
		if v.LogType == LogTypeJson {
			n, err = v.OutInput.GetLogWriter().Write(b)
			if err != nil {
				return n, err
			}
		}
	}
	return 0, nil
}

func (lb *LogContainer) wText(b []byte) (n int, err error) {
	if lb.Opt.OutInput == nil {
		return 0, nil
	}
	for _, v := range lb.Opt.OutInput {
		if v.LogType == LogTypeText {
			n, err = v.OutInput.GetLogWriter().Write(b)
			if err != nil {
				return n, err
			}
		}
	}
	return 0, nil
}

func (lb *LogContainer) rangeLog(f func(lg FullLogger)) {
	if lb.jsonLog != nil {
		f(lb.jsonLog)
	}
	if lb.textLog != nil {
		f(lb.textLog)
	}
}

func (lb *LogContainer) SetLevel(lv Level) {
	lb.rangeLog(func(lg FullLogger) {
		lg.SetLevel(lv)
	})
}

func (lb *LogContainer) SetOutput(iow io.Writer) {
	lb.rangeLog(func(lg FullLogger) {
		lg.SetOutput(iow)
	})
}

func (lb *LogContainer) DefaultLogger() *LogContainer {
	return defLog
}

func (lb *LogContainer) SetLogger(v *LogContainer) {
	defLog = v
}

func (lb *LogContainer) Clone() *LogContainer {
	return &LogContainer{
		Opt:     defLog.Opt,
		jsonLog: defLog.jsonLog,
		textLog: defLog.textLog,
		LogLv:   defLog.LogLv,
	}
}

// Error calls the default logger's Error method.
func (lb *LogContainer) Error(v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Error(v...)
	})
}

// 兼容原来接口 用不着的时候删除
func (lb *LogContainer) Errorln(v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Error(v...)
	})
}

func (lb *LogContainer) Fatal(v ...any) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Error(v...)
	})
	os.Exit(0)
}

// Warn calls the default logger's Warn method.
func (lb *LogContainer) Warn(v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Warn(v...)
	})
}

// Info calls the default logger's Info method.
func (lb *LogContainer) Info(v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Info(v...)
	})
}

// 兼容原来接口 用不着的时候删除
func (lb *LogContainer) Infoln(v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Info(v...)
	})
}

// Debug calls the default logger's Debug method.
func (lb *LogContainer) Debug(v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Debug(v...)
	})
}
func (lb *LogContainer) Trace(v ...any) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Trace(v...)
	})
}
func (lb *LogContainer) Notice(v ...any) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Notice(v...)
	})
}

// Errorf calls the default logger's Errorf method.
func (lb *LogContainer) Errorf(format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Errorf(format, v...)
	})
}

func (lb *LogContainer) Fatalf(format string, v ...any) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Errorf(format, v...)
	})
	os.Exit(0)
}

// Warnf calls the default logger's Warnf method.
func (lb *LogContainer) Warnf(format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Warnf(format, v...)
	})
}

// Infof calls the default logger's Infof method.
func (lb *LogContainer) Infof(format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Infof(format, v...)
	})
}

// Debugf calls the default logger's Debugf method.
func (lb *LogContainer) Debugf(format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Debugf(format, v...)
	})
}

func (lb *LogContainer) Tracef(format string, v ...any) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Tracef(format, v...)
	})
}
func (lb *LogContainer) Noticef(format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.Noticef(format, v...)
	})
}

// CtxErrorf calls the default logger's CtxErrorf method.
func (lb *LogContainer) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxErrorf(ctx, format, v...)
	})
}
func (lb *LogContainer) CtxFatalf(ctx context.Context, format string, v ...any) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxErrorf(ctx, format, v...)
	})
	os.Exit(0)
}

// CtxWarnf calls the default logger's CtxWarnf method.
func (lb *LogContainer) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxWarnf(ctx, format, v...)
	})
}

// CtxInfof calls the default logger's CtxInfof method.
func (lb *LogContainer) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxInfof(ctx, format, v...)
	})
}

// CtxDebugf calls the default logger's CtxDebugf method.
func (lb *LogContainer) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxDebugf(ctx, format, v...)
	})
}

func (lb *LogContainer) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxTracef(ctx, format, v...)
	})
}

func (lb *LogContainer) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	lb.rangeLog(func(lg FullLogger) {
		lg.CtxNoticef(ctx, format, v...)
	})
}
