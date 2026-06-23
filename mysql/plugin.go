package mysql

import (
	"fmt"
	"gorm.io/gorm"
)

var usePlugins = make(map[string]gorm.Plugin)

func SetPlugin(plugin gorm.Plugin) {
	usePlugins[plugin.Name()] = plugin
}

var _ gorm.Plugin = &Plugin{}

type gormHookFunc func(tx *gorm.DB)

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}

type Plugin struct {
}

func (p Plugin) Name() string {
	return "MyPlugin"
}

func (p Plugin) Initialize(db *gorm.DB) (err error) {
	cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		// {cb.Create().Before("gorm:create"), p.before("gorm.Create"), "before:create"},
		{cb.Create().After("gorm:create"), p.after(), "after:create"},

		// {cb.Query().Before("gorm:query"), p.before("gorm.Query"), "before:select"},
		{cb.Query().After("gorm:query"), p.after(), "after:select"},

		// {cb.Delete().Before("gorm:delete"), p.before("gorm.Delete"), "before:delete"},
		{cb.Delete().After("gorm:delete"), p.after(), "after:delete"},

		// {cb.Update().Before("gorm:update"), p.before("gorm.Update"), "before:update"},
		{cb.Update().After("gorm:update"), p.after(), "after:update"},

		// {cb.Row().Before("gorm:row"), p.before("gorm.Row"), "before:row"},
		{cb.Row().After("gorm:row"), p.after(), "after:row"},

		// {cb.Raw().Before("gorm:raw"), p.before("gorm.Raw"), "before:raw"},
		{cb.Raw().After("gorm:raw"), p.after(), "after:raw"},
	}

	var firstErr error

	for _, h := range hooks {
		if err := h.callback.Register("MyPlugin:"+h.name, h.hook); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}

	return firstErr
}

//func (p Plugin) before(spanName string) gormHookFunc {
//	return func(tx *gorm.DB) {}
//}

func (p Plugin) after() gormHookFunc {
	return func(tx *gorm.DB) {

		//switch tx.Error {
		//case nil,
		//	gorm.ErrRecordNotFound,
		//	driver.ErrSkip,
		//	io.EOF, // end of rows iterator
		//	sql.ErrNoRows:
		//	// ignore
		//default:
		//	globalLogger.CtxErrorf(tx.Statement.Context, "%s ; SQL ERROR: %v", logger2.FileLine(utils.FileWithLineNum()), tx.Statement.SQL.String(), tx.Error)
		//}
	}
}
