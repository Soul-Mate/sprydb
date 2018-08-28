package query

import (
	"fmt"
	"bytes"
	"reflect"
	"sort"
	"github.com/Soul-Mate/sprydb/define"
	"github.com/Soul-Mate/sprydb/mapper"
)

func (g *Grammar) CompileInsert(value interface{}, builder *Builder) (string, []interface{}, error) {
	rv := reflect.ValueOf(value)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Ptr:
		return g.CompileInsert(rv.Elem().Interface(), builder)
	case reflect.Struct:
		if rt.NumField() <= 0 {
			return "", nil, define.InsertStructEmptyError
		}
		table, column, parameter, bindings, err := g.processInsertObject(value, builder)
		if err != nil {
			return "", nil, err
		}
		sqlStr := g.buildInsertSql(table, column, parameter)
		return sqlStr, bindings, nil
	case reflect.Slice:
		switch rt.Elem().Kind() {
		case reflect.Map:
			return g.processInsertMultiMap(false, value, builder)
		case reflect.Struct:
			table, column, parameters, bindings, err := g.processInsertMultiObject(rv, builder)
			if err != nil {
				return "", nil, err
			}
			sqlStr := g.buildMultiInsertSql(table, column, parameters)
			return sqlStr, bindings, nil
		default:
			return "", nil, define.InsertSliceTypeError
		}
	case reflect.Map:
		return g.processInsertMap(false, value, builder)
	default:
		return "", nil, define.UnsupportedInsertTypeError
	}
}

// 处理插入一个struct
func (g *Grammar) processInsertObject(obj interface{}, builder *Builder) (
	table, column, parameter string, bindings []interface{}, err error) {
	var (
		columns   []string
		values    []interface{}
		objMapper *mapper.Mapper
	)

	if objMapper, err = mapper.NewMapper(obj, g.syntax, g.styler); err != nil {
		return
	}

	if err = objMapper.Parse(mapper.PARSE_INSERT); err != nil {
		return
	}

	if builder.tableName == "" {
		builder.tableName = objMapper.GetTable()
	}

	if columns, values = objMapper.GetInsertColumnAndValues(); len(columns) <= 0 {
		return
	}
	table = g.syntax.WrapTable(builder.tableName)
	column = g.syntax.ColumnToString(columns)
	parameter = "(" + g.syntax.ParameterByLenToString(len(values)) + ")"
	bindings = values
	return
}

// 处理插入多个struct
func (g *Grammar) processInsertMultiObject(rf reflect.Value, builder *Builder) (
	table, column string, parameters []string, bindings []interface{}, err error) {

	var (
		parameter string
		values    []interface{}
		rfLen     int
	)
	if rfLen = rf.Len(); rfLen <= 0 {
		return
	}

	for next := 0; next < rfLen; next++ {
		rfByIndex := rf.Index(next)
		v := rfByIndex.Interface()
		table, column, parameter, values, err = g.processInsertObject(v, builder)
		if err != nil {
			continue
		}
		parameters = append(parameters, parameter)
		bindings = append(bindings, values...)
	}
	return
}

// 处理插入 map类型
func (g *Grammar) processInsertMap(pointer bool, value interface{}, builder *Builder) (
	sqlStr string, bindings []interface{}, err error) {
	var (
		table     string
		column    string
		parameter string
		ok        bool
		v         map[string]interface{}
	)
	if builder.tableName == "" {
		err = define.TableNoneError
		return
	}
	table = g.syntax.WrapTable(builder.tableName)

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
			return
		}
	}

	if len(v) <= 0 {
		err = define.InsertMapEmptyError
		return
	}
	column, parameter, bindings = g.parseInsertMap(v)
	sqlStr = g.buildInsertSql(table, column, parameter)
	return
}

// 处理插入slice map 类型
func (g *Grammar) processInsertMultiMap(pointer bool, value interface{}, builder *Builder) (
	sqlStr string, bindings []interface{}, err error) {

	var (
		ok        bool
		table     string
		column    string
		parameter string
		values    []map[string]interface{}
	)
	if builder.tableName == "" {
		err = define.TableNoneError
		return
	}
	table = g.syntax.WrapTable(builder.tableName)

	if pointer {
		switch value.(type) {
		case *[]map[string]interface{}:
			ok = true
		}
		if !ok {
			err = define.InsertPointerSliceMapTypeError
			return
		}
		values = *value.(*[]map[string]interface{})
	} else {
		if values, ok = value.([]map[string]interface{}); !ok {
			err = define.MapTypeError
			return
		}
	}

	if len(values) <= 0 {
		err = define.InsertSliceMapEmptyError
		return
	}
	column, parameter, bindings = g.parseInsertMap(values...)
	sqlStr = g.buildInsertSql(table, column, parameter)
	return
}

// 解析插入的map
func (g *Grammar) parseInsertMap(values ...map[string]interface{}) (column, parameter string, bindings []interface{}) {
	var (
		keys []string
		buf  bytes.Buffer
	)

	// 获取key并进行排序
	for k := range values[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	keysLen := len(keys)
	for _, m := range values {
		for i := 0; i < keysLen; i++ {
			if v, ok := m[keys[i]]; !ok {
				// 后续map中没有值的key, 使用空值代替
				bindings = append(bindings, nil)
			} else {
				bindings = append(bindings, v)
			}
		}
		buf.WriteString("(" + g.syntax.ParameterByLenToString(keysLen) + "),")
	}
	column = g.syntax.ColumnToInsertString(keys)
	parameter = buf.String()[:buf.Len()-1]
	return
}

// 生成一个插入语句的sql
func (g *Grammar) buildInsertSql(table, column, parameters string) string {
	if column == "" {
		return ""
	}
	return fmt.Sprintf("insert into %s (%s) values %s;", table, column, parameters)
}

// 生成多个插入语句的sql
func (g *Grammar) buildMultiInsertSql(table, column string, parameters []string) string {
	if column == "" {
		return ""
	}
	var buf bytes.Buffer
	for i := 0; i < len(parameters); i++ {
		buf.WriteString(parameters[i])
		if i != len(parameters)-1 {
			buf.WriteString(",")
		}
	}
	return fmt.Sprintf("insert into %s (%s) values %s;", table, column, buf.String())
}
