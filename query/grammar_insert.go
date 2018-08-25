package query

import (
	"fmt"
	"sort"
	"bytes"
	"errors"
	"reflect"
	"github.com/Soul-Mate/sprydb/define"
	"github.com/Soul-Mate/sprydb/mapper"
)

func (g *Grammar) CompileInsert(value interface{}, builder *Builder) (sqlStr string, bindings []interface{}, err error) {
	var table, columns, parameters string
	reflectV := reflect.ValueOf(value)
	reflectT := reflectV.Type()
	switch reflectT.Kind() {
	case reflect.Map:
		table, columns, parameters, bindings, err = g.processInsertMapType(false, value, builder)
		if err != nil {
			return
		}
	case reflect.Struct:
		table, columns, parameters, bindings, err = g.processInsertObjectType(reflectV, builder)
		if err != nil {
			return
		}
	case reflect.Slice:
		switch reflectT.Elem().Kind() {
		case reflect.Struct:
			table, columns, parameters, bindings, err = g.processInsertMultiObject(reflectV, builder)
			if err != nil {
				return
			}
		case reflect.Map:
			table, columns, parameters, bindings, err = g.processInsertSliceMapType(false, value, builder)
			if err != nil {
				return
			}
		default:
			err = define.UnsupportedTypeError
			return
		}
	case reflect.Ptr:
		// pointer dereference
		return g.CompileInsert(reflectV.Elem().Interface(), builder)
	default:
		err = define.UnsupportedTypeError
		return
	}
	if columns == "" {
		return
	}
	sqlStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, columns, parameters)
	return
}

// 处理插入 map类型
func (g *Grammar) processInsertMapType(pointer bool, value interface{}, builder *Builder) (table,
columns, parameters string, bindings []interface{}, err error) {

	var (
		ok bool
		v  map[string]interface{}
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
			err = errors.New("The map type like map[string]interface{}")
			return
		}
		v = *value.(*map[string]interface{})
	} else {
		if v, ok = value.(map[string]interface{}); !ok {
			err = errors.New("The map type like map[string]interface{}")
			return
		}
	}
	if len(v) <= 0 {
		return
	}
	columns, parameters, bindings = g.processInsertMap(v)
	return
}

// 处理插入slice map 类型
func (g *Grammar) processInsertSliceMapType(pointer bool, value interface{}, builder *Builder) (
	table, columns, parameters string, bindings []interface{}, err error) {

	var (
		ok     bool
		values []map[string]interface{}
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
			err = errors.New("The map type like []map[string]interface{}")
			return
		}
		values = *value.(*[]map[string]interface{})
	} else {
		if values, ok = value.([]map[string]interface{}); !ok {
			err = errors.New("The map type like []map[string]interface{}")
			return
		}
	}
	if len(values) <= 0 {
		return
	}
	columns, parameters, bindings = g.processInsertMap(values...)
	return
}

// 处理插入map
func (g *Grammar) processInsertMap(values ...map[string]interface{}) (
	columnStr, parameterStr string, bindings []interface{}) {
	var (
		keys []string
		buf  bytes.Buffer
	)
	for k := range values[0] {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	keysLen := len(keys)
	for _, m := range values {
		for _, k := range keys {
			if v, ok := m[k]; ok {
				bindings = append(bindings, v)
			}
		}
		buf.WriteString("(" + g.syntax.ParameterByLenToString(keysLen) + "),")
	}
	columnStr = g.syntax.ColumnToInsertString(keys)
	parameterStr = buf.String()[:buf.Len()-1]
	return
}

// 处理插入一个struct
func (g *Grammar) processInsertObjectType(refV reflect.Value, builder *Builder) (
	table, columnStr, parameterStr string, bindings []interface{}, err error) {

	v := refV.Interface()
	// if the value insert is empty, we will not do anything
	//if reflect.DeepEqual(v, reflect.Zero(refV.Type()).Interface()) {
	//	return
	//}
	var (
		column    []string
		values    []interface{}
		objMapper *mapper.Mapper
	)
	if objMapper, err = mapper.NewMapper(v, g.syntax, g.styler); err != nil {
		return
	}

	if err = objMapper.Parse(); err != nil {
		return
	}

	if builder.tableName == "" {
		builder.tableName = objMapper.GetTable()
	}

	if column, values = objMapper.GetColumnsAndValues(); len(column) <= 0 {
		return
	}
	table = g.syntax.WrapTable(builder.tableName)
	columnStr = g.syntax.ColumnToString(column)
	parameterStr = "(" + g.syntax.ParameterByLenToString(len(values)) + ")"
	bindings = append(bindings, values...)
	return
}

// 处理插入多个struct
func (g *Grammar) processInsertMultiObject(refV reflect.Value, builder *Builder) (
	table, columnStr, parameters string, bindings []interface{}, err error) {

	var (
		buf       bytes.Buffer
		column    []string
		values    []interface{}
		refVLen   int
		objMapper *mapper.Mapper
	)
	if refVLen = refV.Len(); refVLen <= 0 {
		return
	}

	next := 0
	// 找到第一个可以用于插入的struct
	// 提取出column
	for ; next < refVLen; next++ {
		refI := refV.Index(next)
		v := refI.Interface()
		if reflect.DeepEqual(v, reflect.Zero(refI.Type()).Interface()) {
			continue
		}

		if objMapper, err = mapper.NewMapper(v, g.syntax, g.styler); err != nil {
			continue
		}

		if err = objMapper.Parse(); err != nil {
			continue
		}

		if builder.tableName == "" {
			builder.tableName = objMapper.GetTable()
		}

		if column, values = objMapper.GetColumnsAndValues(); len(column) <= 0 {
			continue
		}

		buf.WriteString("(" + g.syntax.ParameterByLenToString(len(values)) + "),")
		bindings = append(bindings, values...)
		break
	}

	for i := next + 1; i < refVLen; i++ {
		refI := refV.Index(i)
		v := refI.Interface()

		if reflect.DeepEqual(v, reflect.Zero(refI.Type()).Interface()) {
			continue
		}

		if objMapper, err = mapper.NewMapper(v, g.syntax, g.styler); err != nil {
			continue
		}

		if err = objMapper.Parse(); err != nil {
			continue
		}

		if values = objMapper.GetValuesByColumns(column); values == nil {
			continue
		}
		buf.WriteString("(" + g.syntax.ParameterByLenToString(len(values)) + "),")
		bindings = append(bindings, values...)
	}
	// slice中所有的struct都是空的
	if buf.Len() <= 0 {
		return
	}
	table = g.syntax.WrapTable(builder.tableName)
	columnStr = g.syntax.ColumnToString(column)
	parameters = buf.String()[:buf.Len()-1]
	return
}
