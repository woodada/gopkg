package main

import (
	"context"
	"time"

	"github.com/woodada/gopkg/ctxspan"
	"github.com/woodada/gopkg/logger"
	"github.com/woodada/gopkg/milvus"
	"github.com/woodada/gopkg/mysql"
	"github.com/woodada/gopkg/redis"
)

func main() {
	time.Local = time.FixedZone("INCHINA", 8*3600)
	logger.InitLogger("/tmp/gopkg-example.log", logger.LevelDebug.String(), 10*logger.LogSizeMb)
	logger.Debug("debug without context")
	logger.Info("info without context")
	logger.Warn("warn without context")
	logger.Error("error without context")
	ctx := context.Background()
	ctx = ctxspan.FillSpanContext(ctx)
	logger.CtxDebugf(ctx, "%s", "debugf with context")
	logger.CtxInfof(ctx, "%s", "infof with context")
	logger.CtxWarnf(ctx, "%s", "warnf with context")
	logger.CtxErrorf(ctx, "%s", "errorf with context")

	checkMysql()
	checkMilvus()
	checkRedis()
}

func checkMysql() {
	mysql.SetLogger(logger.DefaultLogger())
	mysql.InitDB(mysql.Config{})
}

func checkRedis() {
	redis.InitRedis(redis.RedisConfig{})
}

func checkMilvus() {
	milvus.NewClient(milvus.Config{})
}
