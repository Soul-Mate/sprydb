package query

import (
	"bytes"
	"fmt"
	"github.com/Soul-Mate/sprydb/syntax"
	"github.com/Soul-Mate/sprydb/binding"
)

type BuilderJoin struct {
	typ   string
	table string
	alias string
	joins []map[string]interface{}
	*Builder
}

func NewJoin(builder *Builder, table, alias, typ string) *BuilderJoin {
	return &BuilderJoin{
		typ:     typ,
		table:   table,
		alias:   alias,
		Builder: builder,
	}
}

func (b *Builder) Join(table, first, operator, second string) *Builder {
	b.join(table, first, operator, second, "inner join")
	return b
}

func (b *Builder) LeftJoin(table, first, operator, second string) *Builder {
	b.join(table, first, operator, second, "left join")
	return b
}

func (b *Builder) RightJoin(table, first, operator, second string) *Builder {
	b.join(table, first, operator, second, "right join")
	return b
}

func (b *Builder) JoinClosure(table string, closure func(*BuilderJoin)) *Builder {
	b.joinClosure(table, "inner join", closure)
	return b
}

func (b *Builder) LeftJoinClosure(table string, closure func(*BuilderJoin)) *Builder {
	b.joinClosure(table, "left join", closure)
	return b
}

func (b *Builder) RightJoinClosure(table string, closure func(*BuilderJoin)) *Builder {
	b.joinClosure(table, "right join", closure)
	return b
}

func (b *Builder) join(table, first, operator, second, typ string) {
	table, alias := b.syntax.ParseTable(table)
	join := NewJoin(NewBuilder(b.driver, b.syntax, binding.NewBinding()), table, alias, typ)
	join.On(first, operator, second)
	b.joins = append(b.joins, join)
	b.joinMap[table] = alias
}

func (b *Builder) joinClosure(table, typ string, closure func(*BuilderJoin)) {
	table, alias := b.syntax.ParseTable(table)
	join := NewJoin(NewBuilder(b.driver, b.syntax, binding.NewBinding()), table, alias, typ)
	b.joins = append(b.joins, join)
	b.joinMap[table] = alias
	closure(join)
}

func (j *BuilderJoin) On(first, operator, second string) *BuilderJoin {
	j.joins = append(j.joins, map[string]interface{}{
		"first":    first,
		"operator": operator,
		"second":   second,
		"logic":    "and",
	})
	return j
}

func (j *BuilderJoin) OrOn(first, operator, second string) *BuilderJoin {
	j.joins = append(j.joins, map[string]interface{}{
		"first":    first,
		"operator": operator,
		"second":   second,
		"logic":    "or",
	})
	return j
}

func (j *BuilderJoin) compile(syntax syntax.Syntax) string {
	var (
		joinLen int
		buf     bytes.Buffer
	)
	joinLen = len(j.joins)
	if joinLen <= 0 {
		return ""
	}
	table := syntax.WrapAliasTable(j.table, j.alias)
	first := syntax.WrapColumn(j.joins[0]["first"].(string))
	operator := j.joins[0]["operator"].(string)
	second := syntax.WrapColumn(j.joins[0]["second"].(string))
	buf.WriteString(
		fmt.Sprintf("%s %s on %s %s %s", j.typ, table, first, operator, second))
	for i := 1; i < joinLen; i++ {
		first = syntax.WrapColumn(j.joins[i]["first"].(string))
		operator = j.joins[i]["operator"].(string)
		second = syntax.WrapColumn(j.joins[i]["second"].(string))
		buf.WriteString(
			fmt.Sprintf(" %s %s %s %s", j.joins[i]["logic"].(string), first, operator, second))
	}
	return buf.String()
}
