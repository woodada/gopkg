package ctxmeta

import (
	"context"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"golang.org/x/text/language"
)

const (
	langKey = "yqfuture-lang"
	codeKey = "yqfuture-code"
	msgKey  = "yqfuture-msg"
)

// GetLangFromContext 从上下文获取语言
func GetLangFromContext(ctx context.Context) string {
	if lang, ok := metainfo.GetPersistentValue(ctx, langKey); ok {
		return lang
	}
	return language.SimplifiedChinese.String()
}

// SetLangToContext 设置语言到上下文
func SetLangToContext(ctx context.Context, lang string) context.Context {
	return metainfo.WithPersistentValue(ctx, langKey, lang)
}

//// SetCodeMsg 设置错误码到本节点上下文
//func SetCodeMsg(ctx context.Context, code int32, msg string) context.Context {
//	ctx = metainfo.WithValue(ctx, codeKey, fmt.Sprint(code))
//	ctx = metainfo.WithValue(ctx, msgKey, msg)
//	return ctx
//}
//
//// GetCodeMsg 设置错误码到本节点上下文
//func GetCodeMsg(ctx context.Context) (int32, string) {
//	var code int32
//	if codeStr, ok := metainfo.GetValue(ctx, codeKey); ok && codeStr != "" {
//		code64, _ := strconv.ParseInt(codeStr, 10, 32)
//		code = int32(code64)
//	}
//	msg, _ := metainfo.GetValue(ctx, msgKey)
//	return code, msg
//}
