package logger

import (
	"os"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/strutil"
	"golang.org/x/exp/slog"
)

type OutInputType string

const (
	OutInputEnvKey              = "LOGGER_SETTING"
	OutInputFile   OutInputType = "FILE"
	OutInputStdout OutInputType = "STDOUT"
)

func InitLogger(logPath string, lvStr string, logSize LogSize) {
	pl := &lOutInput{
		RotationTime: 24 * time.Hour,
		RotationSize: logSize,
		LogPath:      logPath,
	}
	lv := GetLevel(lvStr)

	outInputArr := make([]*LogOutInput, 0)
	outInput := os.Getenv(OutInputEnvKey)
	if outInput != "" {
		oiArr := strutil.SplitEx(outInput, ",", true)
		oiStdout := false
		oiFile := false
		for _, v := range oiArr {
			outMethodAndFormat := strutil.SplitEx(v, ":", true)
			var method string
			if len(outMethodAndFormat) >= 1 {
				method = strings.ToUpper(outMethodAndFormat[0])
			}
			var format string
			if len(outMethodAndFormat) >= 2 {
				format = strings.ToUpper(outMethodAndFormat[1])
			}

			if method == string(OutInputFile) && !oiFile {
				oiFile = true
				outInputArr = append(outInputArr, &LogOutInput{
					LogType:  format,
					OutInput: pl,
				})
			} else if method == string(OutInputStdout) && !oiStdout {
				oiStdout = true
				outInputArr = append(outInputArr, &LogOutInput{
					LogType:  format,
					OutInput: DefLogOutInput,
				})
			}
		}
	}
	if len(outInputArr) == 0 {
		outInputArr = []*LogOutInput{
			{
				LogType:  LogTypeText,
				OutInput: DefLogOutInput,
			},
			{
				LogType:  LogTypeText,
				OutInput: pl,
			},
		}
	}
	lg := New(LogOptions{
		Level:     lv,
		AddSource: true,
		CtxHandle: nil,
		TextColor: true,
		OutInput:  outInputArr,
	})
	SetLogger(lg)
}

func New(opt LogOptions) *LogContainer {
	lb := &LogContainer{
		Opt:     opt,
		jsonLog: nil,
		textLog: nil,
	}
	jsw := NewLogWrite(lb.wJson)
	txw := NewLogWrite(lb.wText)
	lv := &slog.LevelVar{}
	lb.jsonLog = newJsonLog(opt, lv, jsw)
	lb.textLog = newTextLog(opt, lv, txw)
	return lb
}

func newJsonLog(opt LogOptions, lv *slog.LevelVar, lw *LogWrite) FullLogger {
	if len(opt.OutInput) <= 0 {
		return nil
	}
	getJs := func(jo []*LogOutInput) bool {
		for _, v := range jo {
			if v.LogType == LogTypeJson {
				return true
			}
		}
		return false
	}
	if !getJs(opt.OutInput) {
		return nil
	}
	ctxHandle := DefCtxHandle
	if opt.CtxHandle != nil {
		ctxHandle = opt.CtxHandle
	}
	jsonLog := &nLog{
		slog:         newSLog(lw, lv, LogTypeJson, &opt),
		slogLv:       lv,
		ctxHandle:    ctxHandle,
		AddSource:    opt.AddSource,
		SourceOffset: opt.SourceOffset,
		output:       lw,
	}
	jsonLog.SetLevel(opt.Level)
	return jsonLog
}

func newTextLog(opt LogOptions, lv *slog.LevelVar, lw *LogWrite) FullLogger {
	if len(opt.OutInput) <= 0 {
		return nil
	}
	getTx := func(jo []*LogOutInput) bool {
		for _, v := range jo {
			if v.LogType == LogTypeText {
				return true
			}
		}
		return false
	}
	if !getTx(opt.OutInput) {
		return nil
	}
	ctxHandle := DefCtxHandle
	if opt.CtxHandle != nil {
		ctxHandle = opt.CtxHandle
	}
	textLog := &nLog{
		slog:         newSLog(lw, lv, LogTypeText, &opt),
		slogLv:       lv,
		ctxHandle:    ctxHandle,
		AddSource:    opt.AddSource,
		SourceOffset: opt.SourceOffset,
		output:       lw,
	}
	textLog.SetLevel(opt.Level)
	return textLog
}
