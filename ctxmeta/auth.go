package ctxmeta

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
)

const (
	authKey = "yqfeature-auth"
)

// GetAuth 从上下文获取语言
func GetAuth(ctx context.Context) (AuthCtx, bool) {
	authStr, ok := metainfo.GetPersistentValue(ctx, authKey)
	if !ok {
		return AuthCtx{}, false
	}
	a := parseAuthCtx(authStr)
	return a, true
}

// MustGetAuth 从上下文获取语言
func MustGetAuth(ctx context.Context) AuthCtx {
	a, _ := GetAuth(ctx)
	return a
}

// SetAuth 设置授权信息到上下文
func SetAuth(ctx context.Context, authCtx AuthCtx) context.Context {
	return metainfo.WithPersistentValue(ctx, authKey, authCtx.JSON())
}
