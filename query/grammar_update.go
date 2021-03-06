package query

import (
	"reflect"
	"github.com/Soul-Mate/sprydb/define"
	"fmt"
	"bytes"
	"github.com/Soul-Mate/sprydb/mapper"
)

func (g *Grammar) CompileUpdate(value interface{}, builder *Builder) (string, []interface{}, error) {
	var (
		err                           error
		bindings                      []interface{}
		table, columns, wheres, joins string
	)
	v := reflect.ValueOf(value)
	t := v.Type()
	switch t.Kind() {
	case reflect.Map:
		table, columns, bindings, err = g.processUpdateMapType(false, value, builder)
		if err != nil {
			return "", nil, nil
		}
	case reflect.Struct:
		table, columns, bindings, err = g.processUpdateObjectType(value, builder)
		if err != nil {
			return "", nil, err
		}
	case reflect.Ptr:
		return g.CompileUpdate(v.Elem().Interface(), builder)
	default:
		return "", nil, define.UnsupportedUpdateTypeError
	}
	if columns == "" {
		return "", nil, nil
	}
	wheres = g.CompileWhere(builder.wheres, true)
	joins = g.CompileJoin(builder.joins)
	sqlStr := fmt.Sprintf("update %s %s set %s %s", table, joins, columns, wheres)
	return sqlStr, bindings, nil
}

func (g *Grammar) processUpdateMapType(pointer bool, value interface{}, builder *Builder) (
	table, column string, bindings []interface{}, err error) {
	var (
		ok bool
		v  map[string]interface{}
	)

	if builder.tableName == "" {
		err = define.TableNoneError
		return
	}

	if pointer {
		switch value.(type) {
		case *map[string]interface{}:
			ok = true
		}
		if !ok {
			err = define.PointerMapTypeError
			return
		}
		v = *value.(*map[string]interface{})
	} else {
		if v, ok = value.(map[string]interface{}); !ok {
			err = define.MapTypeError
		}
	}

	if len(v) <= 0 {
		err = define.UpdateEmptyMapError
		return
	}

	buf := bytes.Buffer{}
	for mk, mv := range v {
		buf.WriteString(g.syntax.WrapColumn(mk))
		buf.WriteString(" = ?,")
		bindings = append(bindings, mv)
	}
	table = g.syntax.WrapAliasTable(builder.tableName, builder.tableAlias)
	column = buf.String()[:buf.Len()-1]
	return
}

func (g *Grammar) processUpdateObjectType(value interface{}, builder *Builder) (
	table, column string, bindings []interface{}, err error) {
	var (
		columns   []string
		values    []interface{}
		objMapper *mapper.Mapper
	)
	if objMapper, err = mapper.NewMapper(value, g.syntax, g.styler); err != nil {
		return
	}

	if builder.tableName != "" {
		objMapper.SetTable(builder.tableName)
		objMapper.SetAlias(builder.tableAlias)
	}

	objMapper.SetJoinMap(&builder.joinMap)
	if err = objMapper.Parse(mapper.PARSE_UPDATE); err != nil {
		return
	}

	if builder.tableName == "" {
		builder.tableName = objMapper.GetTable()
		builder.tableAlias = objMapper.GetAlias()
	}

	if columns, values = objMapper.GetUpdateColumnAndValues(); len(columns) <= 0 {
		err = define.UpdateEmptyStructError
		return
	}

	table = g.syntax.WrapAliasTable(builder.tableName, builder.tableAlias)
	column = g.syntax.ColumnToUpdateString(columns)
	bindings = append(bindings, values...)
	return
}
