package mapper

import (
	"reflect"
	"github.com/Soul-Mate/sprydb/syntax"
)

type TableMapper struct {
	table   string
	alias   string
	t       *reflect.Type
	v       *reflect.Value
	syntax  syntax.Syntax
	joinMap map[string]string
}

func NewTableMapper(t *reflect.Type, v *reflect.Value, syntax syntax.Syntax) *TableMapper {
	return &TableMapper{
		t:      t,
		v:      v,
		syntax: syntax,
	}
}
// 解析table和alias
// 当用户没有调用Table()时
func (tm *TableMapper) parse(parserFun func(string) string) {
	if tm.table == "" {
		// 调用用户为struct定义的Table()方法
		if table := tm.callTableMethod(); table != "" {
			tm.table, tm.alias = tm.syntax.ParseTable(table)
		} else {
			// 用户没有定义方法,解析struct的名称
			fName := (*tm.t).Name()
			if fName != "" {
				tm.table = parserFun(fName)
				return
			}
		}
	}
}

// 调用映射struct的Table方法
func (tm *TableMapper) callTableMethod() string {
	method := tm.v.MethodByName("Table")
	if !method.IsValid() {
		return ""
	}
	ret := method.Call(nil)[0]
	if ret.Kind() != reflect.String {
		return ""
	}
	return ret.String()
}

func (tm *TableMapper) GetTable() string {
	return tm.table
}

func (tm *TableMapper) SetTable(table string) {
	tm.table = table
}

func (tm *TableMapper) GetAlias() string {
	return tm.alias
}

func (tm *TableMapper) SetAlias(alias string) {
	tm.alias = alias
}

func (tm *TableMapper) SetJoinMap(aliasMap map[string]string) {
	tm.joinMap = aliasMap
}
