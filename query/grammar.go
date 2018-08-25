package query

import (
	"bytes"
	"fmt"
	"strings"
	"github.com/Soul-Mate/sprydb/define"
	"github.com/Soul-Mate/sprydb/syntax"
	"github.com/Soul-Mate/sprydb/binding"
	"github.com/Soul-Mate/sprydb/mapper"
)

type GrammarInterface interface {
	CompileSelect(builder *Builder) (string, error)
	CompileFind(distinct bool, columns []string, table, alias, pk string) string
	CompileColumns(distinct bool, columns []string) string
	CompileFrom(table, alias string) string
	CompileJoin(joins []*BuilderJoin) string
	CompileWhere(wheres []map[string]interface{}, removeLeading bool) string
	CompileOrderBy(orders map[string]interface{}) string
	CompileInsert(value interface{}, builder *Builder) (sqlStr string, bindings []interface{}, err error)
	CompileUpdate(value interface{}, builder *Builder) (string, []interface{}, error)
	CompileDelete(builder *Builder) (sqlStr string, err error)
}

func NewGrammarFactory(driver string, syntax syntax.Syntax, binding *binding.Binding, styler mapper.MapperStyler) GrammarInterface {
	switch driver {
	case "mysql":
		return NewMysqlGrammar(syntax, binding, styler)
	default:
		return nil
	}
}

type Grammar struct {
	syntax       syntax.Syntax
	styler       mapper.MapperStyler
	binding      *binding.Binding
	selectSqlMap map[string]string
}

var SelectStep = []string{
	"column", "from", "join", "where", "orderBy", "offset",
}

func NewGrammar(syntax syntax.Syntax, binding *binding.Binding, styler mapper.MapperStyler) *Grammar {
	return &Grammar{
		syntax:       syntax,
		binding:      binding,
		styler:       styler,
		selectSqlMap: make(map[string]string),
	}
}

func (g *Grammar) CompileSelect(builder *Builder) (string, error) {
	var (
		column, from, join, where, order, offset string
		buf                                      bytes.Buffer
	)
	if builder.tableName == "" {
		return "", define.TableNoneError
	}
	column = g.CompileColumns(builder.distinct, builder.column)
	from = g.CompileFrom(builder.tableName, builder.tableAlias)
	join = g.CompileJoin(builder.joins)
	where = g.CompileWhere(builder.wheres, true)
	order = g.CompileOrderBy(builder.orders)
	offset = g.CompileOffset(builder.limit, builder.offset)
	g.selectSqlMap["column"] = column
	g.selectSqlMap["from"] = from
	g.selectSqlMap["join"] = join
	g.selectSqlMap["where"] = where
	g.selectSqlMap["orderBy"] = order
	g.selectSqlMap["offset"] = offset
	for i, n := 0, len(SelectStep); i < n; i++ {
		if g.selectSqlMap[SelectStep[i]] != "" {
			buf.WriteString(g.selectSqlMap[SelectStep[i]])
			buf.WriteString(" ")
		}
	}
	return buf.String()[:buf.Len()-1], nil
}

// compile select  statement
func (g *Grammar) CompileColumns(distinct bool, columns []string) string {
	var selectStr, columnStr string
	if len(columns) <= 0 {
		columnStr = "*"
	} else {
		columnStr = g.syntax.ColumnToString(columns)
	}
	if distinct {
		selectStr = "select distinct "
	} else {
		selectStr = "select "
	}
	return selectStr + columnStr
}

// compile from table statement
func (g *Grammar) CompileFrom(table, alias string) string {
	var fromTable string
	if table != "" && alias != "" {
		fromTable = g.syntax.WrapAliasTable(table, alias)
	} else {
		fromTable = g.syntax.WrapTable(table)
	}
	return "from " + fromTable
}

// compile join statement
func (g *Grammar) CompileJoin(joins []*BuilderJoin) string {
	buf := bytes.Buffer{}
	for _, j := range joins {
		buf.WriteString(g.processJoin(j))
		buf.WriteString(" ")
		// merge binding
		g.binding.AddBinding("where", j.binding.GetBindings())
	}
	bufLen := buf.Len()
	if bufLen <= 0 {
		return ""
	}
	return buf.String()[:bufLen-1]
}

func (g *Grammar) processJoin(j *BuilderJoin) string {
	var (
		joinLen int
		buf     bytes.Buffer
	)
	joinLen = len(j.joins)
	if joinLen <= 0 {
		return ""
	}
	buf.WriteString(j.typ + " ")
	buf.WriteString(g.syntax.WrapAliasTable(j.table, j.alias))
	for i := 0; i < joinLen; i++ {
		logic := j.joins[i]["logic"].(string)
		first := g.syntax.WrapColumn(j.joins[i]["first"].(string))
		operator := j.joins[i]["operator"].(string)
		second := g.syntax.WrapColumn(j.joins[i]["second"].(string))
		if i == 0 {
			logic = "on"
		}
		buf.WriteString(
			fmt.Sprintf(" %s %s %s %s", logic, first, operator, second))
	}
	where := g.CompileWhere(j.wheres, false)
	if where != "" {
		buf.WriteString(" ")
		buf.WriteString(where)
	}
	return buf.String()
	//table := g.syntax.WrapAliasTable(j.table, j.alias)
	//first := g.syntax.WrapColumn(j.joins[0]["first"].(string))
	//operator := j.joins[0]["operator"].(string)
	//second := g.syntax.WrapColumn(j.joins[0]["second"].(string))
	//buf.WriteString(
	//	fmt.Sprintf("%s %s on %s %s %s", j.typ, table, first, operator, second))
	//for i := 1; i < joinLen; i++ {
	//	first = g.syntax.WrapColumn(j.joins[i]["first"].(string))
	//	operator = j.joins[i]["operator"].(string)
	//	second = g.syntax.WrapColumn(j.joins[i]["second"].(string))
	//	buf.WriteString(
	//		fmt.Sprintf(" %s %s %s %s", j.joins[i]["logic"].(string), first, operator, second))
	//}
	//return buf.String()
}

// compile where statement
func (g *Grammar) CompileWhere(wheres []map[string]interface{}, removeLeading bool) string {
	var whereSlice []string
	if len(wheres) <= 0 {
		return ""
	}
	for _, where := range wheres {
		switch where["type"] {
		case "Basic":
			if str := g.whereBasic(where); str != "" {
				whereSlice = append(whereSlice, str)
			}
		case "In":
			if str := g.whereIn(where); str != "" {
				whereSlice = append(whereSlice, str)
			}
		case "Between":
			if str := g.whereBetween(where); str != "" {
				whereSlice = append(whereSlice, str)
			}
		case "Null":
			if str := g.whereNull(where); str != "" {
				whereSlice = append(whereSlice, str)
			}
		case "Sub":
			if str := g.whereSub(where); str != "" {
				whereSlice = append(whereSlice, str)
			}
		}
	}
	sqlStr := strings.Join(whereSlice, " ")
	if sqlStr != "" && removeLeading {
		return "where " + removeWhereLeading(sqlStr)
	}
	return sqlStr
}

// where base statement
func (g *Grammar) whereBasic(where map[string]interface{}) string {
	placeholder := g.syntax.ParameterByInterfaceToString(where["value"])
	if placeholder == "" {
		return ""
	}
	logic := where["logic"].(string)
	column := g.syntax.WrapColumn(where["column"].(string))
	operator := where["operator"].(string)
	//g.wheres = append(g.wheres,
	//	fmt.Sprintf("%s %s %s %s", logic, column, operator, placeholder))
	return fmt.Sprintf("%s %s %s %s", logic, column, operator, placeholder)
}

// where in statement
func (g *Grammar) whereIn(where map[string]interface{}) string {
	var typ string
	placeholder := g.syntax.ParameterByInterfaceToString(where["value"])
	if placeholder == "" {
		// TODO error no support type placeholder
		return ""
	}
	println(placeholder)
	logic := where["logic"].(string)
	column := g.syntax.WrapColumn(where["column"].(string))
	if not := where["not"].(bool); not {
		typ = "not in"
	} else {
		typ = "in"
	}
	// TODO column slice support
	//g.wheres = append(g.wheres,
	//	fmt.Sprintf("%s %s %s (%s)", logic, column, typ, placeholder))
	return fmt.Sprintf("%s %s %s (%s)", logic, column, typ, placeholder)
}

// where between statement
func (g *Grammar) whereBetween(where map[string]interface{}) string {
	var typ string
	logic := where["logic"].(string)
	column := g.syntax.WrapColumn(where["column"].(string))
	if not := where["not"].(bool); not {
		typ = "not between"
	} else {
		typ = "between"
	}
	//g.wheres = append(g.wheres, fmt.Sprintf("%s %s %s ? and ?", logic, column, typ))
	return fmt.Sprintf("%s %s %s ? and ?", logic, column, typ)
}

// where null statement
func (g *Grammar) whereNull(where map[string]interface{}) string {
	var typ string
	logic := where["logic"].(string)
	column := g.syntax.WrapColumn(where["column"].(string))
	if not := where["not"].(bool); not {
		typ = "is not null"
	} else {
		typ = "is null"
	}
	//g.wheres = append(g.wheres, fmt.Sprintf("%s %s %s", logic, column, typ))
	return fmt.Sprintf("%s %s %s", logic, column, typ)
}

// where sub query statement
func (g *Grammar) whereSub(where map[string]interface{}) string {
	column := g.syntax.WrapColumn(where["column"].(string))
	operator := where["operator"].(string)
	logic := where["logic"].(string)
	builder := where["builder"].(*Builder)
	subSelect, err := g.CompileSelect(builder)
	if err != nil {
		return ""
	}
	// merge sub query binding
	g.binding.AddBinding("where", builder.binding.GetBindings())
	//g.wheres = append(g.wheres, fmt.Sprintf("%s %s %s (%s)", logic, column, operator, subSelect))
	return fmt.Sprintf("%s %s %s (%s)", logic, column, operator, subSelect)
}

// compile order statement
func (g *Grammar) CompileOrderBy(orders map[string]interface{}) string {
	column := orders["column"].([]string)
	if len(column) <= 0 {
		return ""
	}
	direction := orders["direction"]
	columnStr := g.syntax.ColumnToString(column)
	return fmt.Sprintf("order by %s %s", columnStr, direction)
}

// compile limit offset statement
func (g *Grammar) CompileOffset(limit, offset string) string {
	if offset == "" {
		return ""
	}
	if limit != "" {
		return limit + " " + offset
	}
	return offset
}

func (g *Grammar) CompileFind(distinct bool, columns []string, table, alias, pk string) string {
	return fmt.Sprintf("%s from %s where %s = ?",
		g.CompileColumns(distinct, columns),
		g.syntax.WrapAliasTable(table, alias),
		g.syntax.WrapColumn(pk),
	)
}

func removeWhereLeading(s string) string {
	if s[:3] == "or " {
		return s[3:]
	} else {
		return s[4:]
	}
}
