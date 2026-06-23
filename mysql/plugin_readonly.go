package mysql

/* Example
readonlyPlugin := mysql.NewReadonlyPlugin((&model.KfgptTask{}).TableName())
mysql.SetPlugin(readonlyPlugin)
*/

import (
	"fmt"
	"gorm.io/gorm"
)

var _ gorm.Plugin = &ReadonlyPlugin{}

type ReadonlyPlugin struct {
	tableMap map[string]struct{}
}

func NewReadonlyPlugin(tableName ...string) *ReadonlyPlugin {
	tableMap := make(map[string]struct{}, len(tableName))
	for _, name := range tableName {
		tableMap[name] = struct{}{}
	}
	return &ReadonlyPlugin{tableMap: tableMap}
}

func (p *ReadonlyPlugin) Name() string {
	return "ReadonlyPlugin"
}

func (p *ReadonlyPlugin) Initialize(db *gorm.DB) (err error) {
	cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		{cb.Create().Before("gorm:create"), p.before("gorm.Create"), "before:create"},
		// {cb.Create().After("gorm:create"), p.after(), "after:create"},

		// {cb.Query().Before("gorm:query"), p.before("gorm.Query"), "before:select"},
		// {cb.Query().After("gorm:query"), p.after(), "after:select"},

		{cb.Delete().Before("gorm:delete"), p.before("gorm.Delete"), "before:delete"},
		// {cb.Delete().After("gorm:delete"), p.after(), "after:delete"},

		{cb.Update().Before("gorm:update"), p.before("gorm.Update"), "before:update"},
		// {cb.Update().After("gorm:update"), p.after(), "after:update"},

		// {cb.Row().Before("gorm:row"), p.before("gorm.Row"), "before:row"},
		// {cb.Row().After("gorm:row"), p.after(), "after:row"},

		// {cb.Raw().Before("gorm:raw"), p.before("gorm.Raw"), "before:raw"},
		// {cb.Raw().After("gorm:raw"), p.after(), "after:raw"},
	}

	var firstErr error

	for _, h := range hooks {
		if err := h.callback.Register("ReadonlyPlugin:"+h.name, h.hook); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}

	return firstErr
}

func (p *ReadonlyPlugin) before(spanName string) gormHookFunc {
	return func(tx *gorm.DB) {
		if _, ok := p.tableMap[tx.Statement.Table]; !ok {
			panic(tx.Statement.Table + " Readonly")
		}
	}
}

func (p *ReadonlyPlugin) after() gormHookFunc {
	return func(tx *gorm.DB) {

	}
}
