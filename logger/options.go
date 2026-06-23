package logger

type LogType = string

const (
	LogTypeJson LogType = "JSON"
	LogTypeText LogType = "TEXT"
)

type LogOptions struct {
	Level        Level
	AddSource    bool
	SourceOffset int
	TextColor    bool
	Attr         []Attr
	CtxHandle    CtxHandle
	OutInput     []*LogOutInput
	MaxLogSize   LogSize
}

type LogOutInput struct {
	LogType  LogType
	OutInput OutInputLog
}
