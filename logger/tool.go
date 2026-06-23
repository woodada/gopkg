package logger

import (
	"strings"
)

func SpiltFilePath(filePath string) string {
	filePath = strings.ReplaceAll(strings.Trim(filePath, " "), "\\", "/")
	strArr := strings.Split(filePath, "/")
	if len(strArr) > 4 {
		strArr = strArr[len(strArr)-4:]
		filePath = "/" + strings.Join(strArr, "/")
	}
	return filePath
}
