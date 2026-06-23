package logger

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/fatih/color"
	"golang.org/x/exp/slog"
	"io"
	"strings"
	"sync"
)

type NlogHandle struct {
	opt LogOptions
	slog.Handler
	attrs []Attr
	mu    sync.Mutex
	w     io.Writer
}

func (h *NlogHandle) Handle(ctx context.Context, r slog.Record) error {
	timeStr := ""
	levelStr := ""
	msgStr := ""
	fileStr := ""
	traceId := "00000000000000000000000000000000"
	spanId := "0000000000000000"
	other := ""
	//rep := h.opts.ReplaceAttr
	if !r.Time.IsZero() {
		//key := slog.TimeKey
		//val := r.Time.Round(0) // strip monotonic to match Attr behavior
		timeStr = fmt.Sprintf("%v", r.Time.Format("2006-01-02 15:04:05.000"))
		//if rep == nil {
		//	//state.appendKey(key)
		//	//state.appendTime(val)
		//} else {
		//	//state.appendAttr(Time(key, val))
		//}
	}
	// level
	//key := slog.LevelKey
	val := r.Level
	levelStr = val.String()
	//if rep == nil {
	//	//state.appendKey(key)
	//	//state.appendString(val.String())
	//} else {
	//	//state.appendAttr(Any(key, val))
	//}
	// source
	//if h.opts.AddSource {
	//	//state.appendAttr(Any(SourceKey, r.source()))
	//}
	//key = slog.MessageKey
	msg := r.Message
	msgStr = msg
	//if rep == nil {
	//	//state.appendKey(key)
	//	//state.appendString(msg)
	//} else {
	//	//state.appendAttr(String(key, msg))
	//}
	if len(h.attrs) > 0 {
		for _, v := range h.attrs {
			other += " " + v.String()
		}
	}
	r.Attrs(func(a Attr) bool {
		switch a.Key {
		case LineKey:
			fileStr = a.Value.String()
		case CtxTraceId:
			traceId = a.Value.String()
		case CtxSpanId:
			spanId = a.Value.String()
		default:
			other += " " + a.String()
		}
		return true
	})
	msgArr := strutil.SplitEx(msgStr, "\n", false)
	logMsgArr := make([]string, 0)
	if len(msgArr) > 0 {
		newMsg := ""
		for _, v := range msgArr {
			if strings.Trim(v, " ") != "" {
				logMsgArr = append(logMsgArr, v)
			}
		}
		msgStr = newMsg
	}

	if h.opt.TextColor {
		switch r.Level {
		case slog.LevelDebug:
			levelStr = color.MagentaString(levelStr)
		case slog.LevelInfo:
			levelStr = color.BlueString(levelStr)
		case slog.LevelWarn:
			levelStr = color.YellowString(levelStr)
		case slog.LevelError:
			levelStr = color.RedString(levelStr)
		}
		for k, v := range logMsgArr {
			logMsgArr[k] = color.CyanString(v)
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	var err error
	for k, v := range logMsgArr {
		if k == 0 {
			_, err = h.w.Write([]byte(fmt.Sprintf("%v %-5v %v %v %v %v %v \n", timeStr, levelStr, traceId, spanId, fileStr, v, other)))
		} else {
			_, err = h.w.Write([]byte(fmt.Sprintf("%v %-5v %v %v %v \n", timeStr, levelStr, traceId, spanId, v)))
		}
		if err != nil {
			return err
		}
	}
	_, err = h.w.Write([]byte("\n"))
	return err
}
