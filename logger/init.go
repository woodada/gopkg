package logger

import (
	"golang.org/x/exp/slog"
	"io"
	"sync"
)

func newSLog(w io.Writer, lv *slog.LevelVar, logType LogType, logOpt *LogOptions) *slog.Logger {
	opt := LogOptions{}
	if logOpt != nil {
		opt = *logOpt
	}
	var logHandle slog.Handler
	if logType == LogTypeJson {
		logHandle = slog.NewJSONHandler(w, getHandlerOptions(lv)).WithAttrs(opt.Attr)
	} else {
		//logHandle = slog.NewTextHandler(w, getHandlerOptions(lv)).WithAttrs(opt.Attr)
		logHandle = &NlogHandle{
			opt:     opt,
			Handler: slog.NewTextHandler(w, getHandlerOptions(lv)).WithAttrs(opt.Attr),
			attrs:   opt.Attr,
			mu:      sync.Mutex{},
			w:       w,
		}
	}
	return slog.New(logHandle)
}

func getHandlerOptions(lv *slog.LevelVar) *slog.HandlerOptions {
	handOpt := &slog.HandlerOptions{
		AddSource:   false,
		Level:       lv,
		ReplaceAttr: nil,
	}
	return handOpt
}
