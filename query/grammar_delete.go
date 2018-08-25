package query

import (
	"github.com/Soul-Mate/sprydb/define"
	"fmt"
)

func (g *Grammar) CompileDelete(builder *Builder) (sqlStr string, err error) {
	var wheres string
	if builder.tableName == "" {
		err = define.TableNoneError
		return
	}
	wheres = g.CompileWhere(builder.wheres, true)
	sqlStr = fmt.Sprintf("delete from %s %s", builder.tableName, wheres)
	return
}
