package mapper

import (
	"reflect"
	"github.com/Soul-Mate/sprydb/syntax"
)

type tableMapper struct {
	table string
	alias string
	jm    *map[string]string // join map table -> alias
}

func (tm *tableMapper) Parse(ov reflect.Value, ot reflect.Type, syntax syntax.Syntax, style MapperStyler) {
	if tm.table == "" {
		table := CallTableMethod(ov)
		if table != "" {
			if alias, ok := tm.findAliasInJoinMap(table); ok {
				tm.alias = alias
				tm.table = table
			} else {
				tm.table, tm.alias = syntax.ParseTable(table)
			}
		} else {
			if name := ot.Name(); name != ""{
				tm.table = style.table(name)
			}
		}
	}
}

// 根据table在join map查找指定的alias
func (tm *tableMapper) findAliasInJoinMap(table string) (string, bool) {
	if tm.jm == nil {
		return "", false
	}
	alias, ok := (*tm.jm)[table]
	return alias, ok
}
