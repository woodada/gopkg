package milvus

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var milvusSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	milvusSourceDir = sourceDir(file)
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "/ai-server-go/pkg/milvus" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	pcs := [13]uintptr{}
	// the third caller usually from gorm internal
	len := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:len])
	for i := 0; i < len; i++ {
		// second return value is "more", not "ok"
		frame, _ := frames.Next()
		if (!strings.HasPrefix(frame.File, milvusSourceDir) ||
			strings.HasSuffix(frame.File, "_test.go")) && !strings.HasSuffix(frame.File, ".gen.go") {
			return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
		}
	}

	return ""
}
