// Package mysql  init db connection
package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"

	logger2 "github.com/woodada/gopkg/logger"
)

func InitDB(c Config) (*gorm.DB, error) {
	const dsnFormat = "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf(dsnFormat, c.User, c.Password, c.Host, c.Port, c.DB)
	newLogger := newDbLogger()

	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NameReplacer:  strings.NewReplacer("Model", ""),
		},
		Logger:         newLogger,
		TranslateError: true,
		Plugins:        usePlugins,
	})

	if err != nil {
		logger2.Fatalf("init mysql %s failed: %s", dsn, err.Error())
		return nil, err
	}
	return db, nil
}

type dbLogger struct {
	logger Logger
	level  logger.LogLevel
}

func newDbLogger() *dbLogger {
	return &dbLogger{
		logger: globalLogger,
		level:  logger.Info,
	}
}

func (l dbLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l dbLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if l.level >= logger.Info {
		l.logger.CtxInfof(ctx, s, i...)
	}
}

func (l dbLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.level >= logger.Warn {
		l.logger.CtxWarnf(ctx, s, i...)
	}
}

func (l dbLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.level >= logger.Error {
		l.logger.CtxErrorf(ctx, s, i...)
	}
}

const warnElapsedMillisecond = 5000 // SQL执行超过5秒告警
// Trace print sql message
func (l dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin).Milliseconds()
	sql, rows := fc()
	if len(sql) > 4000 {
		sql = sql[:4000] + " ...(truncated)"
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrDuplicatedKey) {
			l.logger.CtxWarnf(ctx, "sql: %s rows: %d useTime: %d error: %v", logger2.FileLine(utils.FileWithLineNum()), sql, rows, elapsed, err)
		} else {
			l.logger.CtxErrorf(ctx, "sql: %s rows: %d useTime: %d error: %v", logger2.FileLine(utils.FileWithLineNum()), sql, rows, elapsed, err)
		}
		return
	}

	if elapsed > warnElapsedMillisecond {
		l.logger.CtxWarnf(ctx, "sql: %s rows: %d useTime: %d SLOW", logger2.FileLine(utils.FileWithLineNum()), sql, rows, elapsed)
		return
	}

	if l.level <= logger.Silent {
		return
	}
	// 减少日志打印
	l.logger.CtxDebugf(ctx, "sql: %s rows: %d useTime: %d", logger2.FileLine(utils.FileWithLineNum()), sql, rows, elapsed)
}
