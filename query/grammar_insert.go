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
	rv := reflect.ValueOf(value)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Struct:
		table, columns, parameters, bindings, err = g.processInsertObjectType(value, builder)
		if err != nil {
			return
		}
	case reflect.Slice:
		if rt.Elem().Kind() != reflect.Struct {
			err = define.InserSliceTypeError
			return
		}
		g.processInsertMultiObject(rv, builder)

	}
	if rt.Kind() != reflect.Ptr {
		err = define.InsertPointerTypeError
		return
	}

	if rt.Elem().Kind() != reflect.Struct {
		err = define.InsertPointerDeferenceTypeError
		return
	}
	table, columns, parameters, bindings, err = g.processInsertObjectType(value, builder)
	if err != nil {
		return
	}
	sqlStr = g.buildInsertSql(table, columns, parameters)
	return

	//switch rt.Elem().Kind() {
	//case reflect.Map:
	//	table, columns, parameters, bindings, err = g.processInsertMapType(false, value, builder)
	//	if err != nil {
	//		return
	//	}
	//case reflect.Struct:
	//	table, columns, parameters, bindings, err = g.processInsertObjectType(value, builder)
	//	if err != nil {
	//		return
	//	}
	//case reflect.Slice:
	//	switch rt.Elem().Kind() {
	//	case reflect.Struct:
	//		table, columns, parameters, bindings, err = g.processInsertMultiObject(rv, builder)
	//		if err != nil {
	//			return
	//		}
	//	case reflect.Map:
	//		table, columns, parameters, bindings, err = g.processInsertSliceMapType(false, value, builder)
	//		if err != nil {
	//			return
	//		}
	//	default:
	//		err = define.UnsupportedTypeError
	//		return
	//	}
	//case reflect.Ptr:
	//	// pointer dereference
	//	return g.CompileInsert(rv.Elem().Interface(), builder)
	//default:
	//	err = define.UnsupportedTypeError
	//	return
	//}
	//if columns == "" {
	//	return
	//}
	//sqlStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, columns, parameters)
	//return
}

func (g *Grammar) CompileInsertMulti(builder *Builder, objects ...interface{}) (
	sqlStr string, bindings []interface{}, err error) {
	var (
		values                   []interface{}
		parameters               []string
		table, column, parameter string
	)
	if len(objects) <= 0 {
		err = define.MultiInsertNoObjectError
		return
	}

	table, column, parameter, values, err = g.processInsertObjectType(objects[0], builder)
	if err != nil {
		return
	}

	parameters = append(parameters, parameter)
	bindings = append(bindings, values...)
	for i := 1; i < len(objects); i++ {
		parameter, values, err = g.getInsertObjectParameterAndBindings(objects[i])
		if err != nil {
			return
		}
		parameters = append(parameters, parameter)
		bindings = append(bindings, values...)
	}
	sqlStr = g.buildMultiInsertSql(table, column, parameters)
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
func (g *Grammar) processInsertObjectType(obj interface{}, builder *Builder) (
	table, columnStr, parameterStr string, bindings []interface{}, err error) {
	var (
		column    []string
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

	if column, values = objMapper.GetInsertColumnAndValues(); len(column) <= 0 {
		return
	}
	table = g.syntax.WrapTable(builder.tableName)
	columnStr = g.syntax.ColumnToString(column)
	parameterStr = "(" + g.syntax.ParameterByLenToString(len(values)) + ")"
	bindings = values
	return
}

func (g *Grammar) getInsertObjectParameterAndBindings(object interface{}) (
	parameter string, bindings []interface{}, err error) {

	var objMapper *mapper.Mapper
	if objMapper, err = mapper.NewMapper(object, g.syntax, g.styler); err != nil {
		return
	}

	if err = objMapper.Parse(mapper.PARSE_INSERT); err != nil {
		return
	}

	bindings = objMapper.GetInsertValues()
	parameter = "(" + g.syntax.ParameterByLenToString(len(bindings)) + ")"
	return
}

// 处理插入多个struct
func (g *Grammar) processInsertMultiObject(rf reflect.Value, builder *Builder) (table, column, parameter string, bindings []interface{}, err error) {

	var (
		//buf        bytes.Buffer
		parameters []string
		values     []interface{}
		rfLen      int
		//objMapper  *mapper.Mapper
	)
	if rfLen = rf.Len(); rfLen <= 0 {
		return
	}

	//next := 0
	// 找到第一个可以用于插入的struct
	// 提取出column
	for next := 0; next < rfLen; next++ {
		rfByIndex := rf.Index(next)
		v := rfByIndex.Interface()
		table, column, parameter, values, err = g.processInsertObjectType(v, builder)
		if err != nil {
			continue
		}
		parameters = append(parameters, parameter)
		bindings = append(bindings, values...)
		//if reflect.DeepEqual(v, reflect.Zero(rfByIndex.Type()).Interface()) {
		//	continue
		//}

		//if objMapper, err = mapper.NewMapper(v, g.syntax, g.styler); err != nil {
		//	continue
		//}
		//
		//if err = objMapper.Parse(mapper.PARSE_INSERT); err != nil {
		//	continue
		//}
		//
		//if builder.tableName == "" {
		//	builder.tableName = objMapper.GetTable()
		//}
		//
		//if columns, values = objMapper.GetInsertColumnAndValues(); len(columns) <= 0 {
		//	continue
		//}
		//
		//buf.WriteString("(" + g.syntax.ParameterByLenToString(len(values)) + "),")
		//bindings = append(bindings, values...)
		//break
	}

	sqlstr := g.buildMultiInsertSql(table, column, parameters)
	println(sqlstr)

	//for i := next + 1; i < rfLen; i++ {
	//	refI := rf.Index(i)
	//	v := refI.Interface()
	//
	//	if reflect.DeepEqual(v, reflect.Zero(refI.Type()).Interface()) {
	//		continue
	//	}
	//
	//	if objMapper, err = mapper.NewMapper(v, g.syntax, g.styler); err != nil {
	//		continue
	//	}
	//
	//	if err = objMapper.Parse(mapper.PARSE_INSERT); err != nil {
	//		continue
	//	}
	//
	//	if values = objMapper.GetValuesByColumns(columns); values == nil {
	//		continue
	//	}
	//	buf.WriteString("(" + g.syntax.ParameterByLenToString(len(values)) + "),")
	//	bindings = append(bindings, values...)
	//}
	// slice中所有的struct都是空的
	//if buf.Len() <= 0 {
	//	return
	//}
	//table = g.syntax.WrapTable(builder.tableName)
	//column = g.syntax.ColumnToString(columns)
	//paramter = buf.String()[:buf.Len()-1]
	return
}
