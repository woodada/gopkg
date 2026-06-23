package logger

type LogSize = int64

const (
	LineKey     = "file"
	CtxTraceId  = "trace_id"
	CtxSpanId   = "span_id"
	SecondLevel = "SecondLevel"

	LogSizeBit LogSize = 1
	LogSizeKb  LogSize = LogSizeBit * 1024
	LogSizeMb  LogSize = LogSizeKb * 1024
	LogSizeGb  LogSize = LogSizeMb * 1024
)
