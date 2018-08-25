package query

import (
	"github.com/Soul-Mate/sprydb/binding"
	"github.com/Soul-Mate/sprydb/syntax"
	"strconv"
)

type Builder struct {
	err        error
	driver     string
	distinct   bool
	column     []string
	tableName  string
	tableAlias string
	TableWrap  string
	joins      []*BuilderJoin
	joinMap    map[string]string
	wheres     []map[string]interface{}
	orders     map[string]interface{}
	limit      string
	offset     string
	binding    *binding.Binding
	syntax     syntax.Syntax
}

func NewBuilder(driver string, syntax syntax.Syntax, binding *binding.Binding) *Builder {
	b := new(Builder)
	b.err = nil
	b.driver = driver
	b.distinct = false
	b.tableName = ""
	b.tableAlias = ""
	b.column = []string{}
	b.joins = []*BuilderJoin{}
	b.joinMap = make(map[string]string)
	b.wheres = []map[string]interface{}{}
	b.orders = make(map[string]interface{})
	b.orders["direction"] = ""
	b.orders["column"] = []string{}
	b.limit = ""
	b.offset = ""
	b.binding = binding
	b.syntax = syntax
	return b
}

func (b *Builder) Table(tableName string) *Builder {
	b.tableName, b.tableAlias = b.syntax.ParseTable(tableName)
	b.joinMap[b.tableName] = b.tableAlias
	return b
}

func (b *Builder) Distinct() *Builder {
	b.distinct = true
	return b
}

func (b *Builder) Select(column ...string) *Builder {
	b.column = column
	return b
}

func (b *Builder) Skip(n int) *Builder {
	if n < 0 {
		n = 0
	}
	b.offset = "offset " + strconv.Itoa(n)
	return b
}

func (b *Builder) Take(n int) *Builder {
	if n < 0 {
		return b
	}
	b.limit = "limit " + strconv.Itoa(n)
	return b
}

func (b *Builder) GetTable() string {
	return b.tableName
}

func (b *Builder) GetAlias() string {
	return b.tableAlias
}

func (b *Builder) SetAlias(alias string)  {
	b.tableAlias = alias
}

func (b *Builder) GetColumn() []string {
	return b.column
}

func (b *Builder) GetDistinct() bool {
	return b.distinct
}

func (b *Builder) GetJoinMap() map[string]string {
	return b.joinMap
}

func (b *Builder) GetErr() error {
	return b.err
}