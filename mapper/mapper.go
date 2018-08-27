package mapper

import (
	"reflect"
	"github.com/Soul-Mate/sprydb/define"
		"errors"
	"regexp"
	"github.com/Soul-Mate/sprydb/syntax"
	"time"
	"fmt"
)

const (
	PARSE = iota
	PARSE_SELECT
	PARSE_INSERT
	PARSE_UPDATE
)

type Mapper struct {
	parseType      int
	joinMap        *map[string]string
	table          string
	alias          string
	pk             string
	t              reflect.Type
	v              reflect.Value
	pv             reflect.Value // 当reflect.Kind 是ptr的时候,将保存reflect.Value的两份值,这份保存指针
	ptr            bool
	orderFName     []string // orderFName 是tagFieldMapper的有序键
	fields         []*Field
	syntax         syntax.Syntax
	tagFieldMapper map[string]*Field
	styler         MapperStyler
}

func NewMapper(object interface{}, syntax syntax.Syntax, styler MapperStyler) (*Mapper, error) {
	var (
		err          error
		reflectValue reflect.Value
	)
	mapper := new(Mapper)
	reflectValue = reflect.ValueOf(object)
	switch reflectValue.Kind() {
	case reflect.Ptr:
		mapper.ptr = true
		mapper.v = reflectValue
		mapper.pv = reflectValue.Elem()
		mapper.t = mapper.pv.Type()
		mapper.tagFieldMapper = make(map[string]*Field)
	case reflect.Struct:
		mapper.v = reflectValue
		mapper.t = reflectValue.Type()
		mapper.tagFieldMapper = make(map[string]*Field)
	default:
		return nil, define.UnsupportedTypeError
	}
	// 解析语法
	mapper.syntax = syntax

	// 解析风格
	if styler == nil {
		mapper.styler = &UnderlineMapperStyle{}
	} else {
		mapper.styler = styler
	}

	if err != nil {
		return nil, err
	}
	return mapper, nil
}

func (m *Mapper) GetPK() string {
	if m.pk == "" {
		pk := CallPKMethod(m.v)
		if m.alias != "" {
			m.pk = m.alias + "." + pk
		} else {
			m.pk = pk
		}
	}
	return m.pk
}

func (m *Mapper) GetTable() string {
	return m.table
	//if m.table == "" {
	//	// table还是为空,则进行名称解析
	//	if m.table = CallTableMethod(m.fv); m.table == "" {
	//		fName := m.t.Name()
	//		if fName == "" {
	//			m.table = ""
	//			return m.table
	//		}
	//		m.table = m.styler.table(fName)
	//	}
	//}
	//return m.table
}

func (m *Mapper) SetTable(table string) {
	m.table = table
}

func (m *Mapper) GetAlias() string {
	return m.alias
}

func (m *Mapper) SetAlias(alias string) {
	m.alias = alias
}

func (m *Mapper) SetJoinMap(aliasMap *map[string]string) {
	m.joinMap = aliasMap
}

func (m *Mapper) GetColumn() (column []string) {
	return m.orderFName
}

func (m *Mapper) GetAddressByColumn(columns []string) (address []interface{}) {
	if len(m.tagFieldMapper) <= 0 {
		return
	}
	for _, c := range columns {
		c = parseColumn(c)

		if f, ok := m.tagFieldMapper[c]; ok {
			address = append(address, f.getSqlNullType())
		}
	}
	return
}

func parseColumn(column string) string {
	re, err := regexp.Compile(`(.*)\s+as\s+(.*)`)
	if err != nil {
		return column
	}
	ss := re.FindStringSubmatch(column)
	if n := len(ss); n >= 3 {
		return ss[1]
	}
	return column
}

func (m *Mapper) GetAddress() (address []interface{}) {
	for _, f := range m.tagFieldMapper {
		address = append(address, f.getSqlNullType())
	}
	return
}

func (m *Mapper) GetColumnAndAddress() (column []string, address []interface{}) {
	// 使用有序的key
	for _, col := range m.orderFName {
		address = append(address, m.tagFieldMapper[col].getSqlNullType())
	}
	column = m.orderFName
	return
}

func recursionGetColumnAndAddress(fs *[]*Field) ([]string, []interface{}, error) {
	var (
		columns    []string
		addresses  []interface{}
		fieldsSize int
	)
	if fieldsSize = len(*fs); fieldsSize <= 0 {
		return nil, nil, define.ObjectNoFieldError
	}
	for i := 0; i < fieldsSize; i++ {
		if (*fs)[i].extend != nil {
			subColumn, subAddress, err := recursionGetColumnAndAddress((*fs)[i].extend.fields)
			if err == nil {
				columns = append(columns, subColumn...)
				addresses = append(addresses, subAddress...)
			}
		} else {
			columns = append(columns, (*fs)[i].tagString)
			addresses = append(addresses, (*fs)[i].getSqlNullType())
		}
	}
	return columns, addresses, nil
}

func (m *Mapper) Parse(parseType int) error {
	var fv reflect.Value
	if m.ptr {
		fv = m.pv
	} else {
		fv = m.v
	}
	m.parseType = parseType
	// 初始化table映射
	m.parseTableMapper(m.v)
	m.fields = m.parseFieldMapper(m.t, fv, m.alias)
	return nil
}

func (m *Mapper) parseTableMapper(fv reflect.Value) {
	// 没有调用session.Table(),
	if m.table == "" {
		// 调用用户为对象定义的table方法
		if table := CallTableMethod(fv); table != "" {
			m.table, m.alias = m.syntax.ParseTable(table)
		} else {
			// 没有定义table方法, 对对象名称进行解析
			// 解析风格可自定义
			fName := m.t.Name()
			if fName != "" {
				m.table = m.styler.table(fName)
				return
			}
		}
	}
}

// 解析struct中的字段
// 如果字段是struct则进根据字段进行递归解析
func (m *Mapper) parseFieldMapper(t reflect.Type, v reflect.Value, alias string) (fields []*Field) {
	var (
		tag *Tag
		fv  reflect.Value
		ff  reflect.StructField
	)
	for i, n := 0, t.NumField(); i < n; i++ {
		ff = t.Field(i)
		fv = v.Field(i)
		// 解析Tag
		if tag = NewTag(&ff, &fv).parse(m.styler.column, m.syntax); tag.ignore {
			continue
		}
		// 处理字段类型
		if field, err := m.processField(ff, fv, tag, alias); err == nil {
			fields = append(fields, field)
		} else {
			panic(err)
		}
	}
	return
}

// 处理字段类型
func (m *Mapper) processField(ff reflect.StructField, fv reflect.Value,
	tag *Tag, alias string) (*Field, error) {
	switch fv.Type().Kind() {
	case
		reflect.Map, reflect.Array,
		reflect.Chan, reflect.Complex64,
		reflect.Func, reflect.Complex128:
		return nil, define.UnsupportedTypeError
	case reflect.Struct:
		return m.structTypeField(ff, fv, tag, false, alias)
	case reflect.Ptr:
		return m.processPtrField(ff, fv, tag, alias)
	default:
		return m.basicTypeField(ff, fv, tag, alias)
	}
}

// 处理指针字段类型
func (m *Mapper) processPtrField(ff reflect.StructField, pfv reflect.Value,
	tag *Tag, alias string) (*Field, error) {
	// 如果是一个空指针并且解析类型是select
	// 那么就对这个空指针进行赋值
	if pfv.IsNil(){
		fmt.Println(tag)
		if m.parseType == PARSE_SELECT {
			if !pfv.CanSet() {
				return nil, define.NullPointerAndNotAssign
			}
			// 对空指针设置内存
			pfv.Set(reflect.New(ff.Type.Elem()))
		}

		// 插入类型为空类型
		if m.parseType == PARSE_INSERT {
			field := &Field{
				tag:       tag,
				addr:      nil,
				typ:       "null",
				tagString: tag.column,
				fv:        nil,
			}
			m.tagFieldMapper[tag.column] = field
			m.orderFName = append(m.orderFName, tag.column)
			return field, nil
		}
	}

	elem := pfv.Elem()
	switch elem.Kind() {
	case
		reflect.Map, reflect.Array,
		reflect.Chan, reflect.Complex64,
		reflect.Func, reflect.Complex128:
		return nil, define.UnsupportedTypeError
	case reflect.Struct:
		return m.structTypeField(ff, pfv, tag, true, alias)
	default:
		return m.basicTypeField(ff, elem, tag, m.alias)
	}
}

// 处理类型是struct的字段以及特殊字段
// 如果ptr是true, ff 和 fv是指针类型
func (m *Mapper) structTypeField(ff reflect.StructField, fv reflect.Value,
	tag *Tag, ptr bool, alias string) (*Field, error) {
	var table string
	addr, err := m.decideColumnAddr(ptr, fv)
	if err != nil {
		return nil, err
	}
	// 用户自定义字段
	if _, ok := addr.(Custom); ok {
		return m.structCustomTypeField(fv, addr, tag, alias), nil
	}

	// 特殊类型字段进行特殊处理
	switch addr.(type) {
	case time.Time, *time.Time: // time字段
		return m.structTimeTypeField(fv, addr, tag, alias), nil
	case Custom: // 用户自定义字段
		return m.structCustomTypeField(fv, addr, tag, alias), nil
	default:
		// 如果是连接查询的struct, 则对alias的值进行判定,
		// 如果不是, 使用最上级的alias
		if tag.extend {
			table, alias = m.decideExtendFieldTableAndAlias(ff, fv, tag)
			// 存在一种情况, 当一个struct中的所有field都是 extend类型时,
			// 并且用户没有为这个struct定义Table, 也没有调用查询Table,
			// 那么就无法得知, 因此需要将第一个extend设置为查询的table
			if m.table == "" {
				m.table, m.alias = table, alias
			}
		}
		// 如果是指针类型,还需要解指针引用
		if ptr {
			fv = fv.Elem()
		}

		// 递归下级字段,将解析的alias传递给struct中的字段
		sub := m.parseFieldMapper(fv.Type(), fv, alias)
		return &Field{
			tag: tag,
			extend: &ExtendField{
				alias:  alias,
				fields: &sub,
			},
		}, nil
	}
	//// Time非地址字段
	//if _, ok := addr.(time.Time); ok {
	//	return m.structTimeTypeField(fv, addr, tag, alias), nil
	//}
	//
	//// Time地址字段
	//if _, ok := addr.(*time.Time); ok {
	//	addr = fv.Interface()
	//	return m.structTimeTypeField(fv, addr, tag, alias), nil
	//}
	// Custom非之针字段 需要获取地址
	//if !ptr {
	//	if fv.CanAddr() {
	//		addr = fv.Addr().Interface()
	//	} else {
	//		addr = fv.Interface()
	//	}
	//
	//	if addr, ok := fv.Addr().Interface().(Custom); ok {
	//		return m.structCustomTypeField(fv, addr, tag, alias), nil
	//	}
	//
	//} else {
	//	addr = fv.Interface()
	//	if _, ok := addr.(Custom); ok {
	//		return m.structCustomTypeField(fv, addr, tag, alias), nil
	//	}
	//}
	//
	// Custom指针字段
	//if _, ok := addr.(Custom); ok {
	//	return m.structCustomTypeField(fv, addr, tag, alias), nil
	//}
	// var pt reflect.Type
	//if ptr {
	//	ft = ff.Type
	//} else {
	//	ft = reflect.PtrTo(ff.Type)
	//}
	//if ft.Implements(reflect.TypeOf((*Custom)(nil)).Elem()) {
	//	var (
	//		column string
	//		iface  interface{}
	//	)
	//	if !tag.useAlias {
	//		column = tag.column
	//	} else {
	//		if alias != "" {
	//			column = alias + "." + tag.column
	//		} else {
	//			column = tag.column
	//		}
	//	}
	//	if ptr {
	//		fv = fv.Elem()
	//	}
	//
	//	if !fv.CanInterface() {
	//		return nil, errors.New("This field cannot get Interface")
	//	}
	//
	//	if fv.CanAddr() {
	//		iface = fv.Addr().Interface()
	//	} else {
	//		iface = fv.Interface()
	//	}
	//
	//	field := &Field{
	//		addr:      iface,
	//		typ:       "custom",
	//		tagString: column,
	//		fv:         &fv,
	//	}
	//	m.tagFieldMapper[column] = field
	//	m.orderFName = append(m.orderFName, column)
	//	return field, nil
	//}
	// 如果是join的struct
	// 定义了alias,则用这个alias覆盖上级的alias
	// 否则会调用定义Table()进行解析
	// 否则统一使用上级的alias
	//if tag.extend && tag.useAlias {
	//	if tag.extendTable != "" {
	//		alias = tag.extendTable
	//	} else {
	//		table, alias = m.decideExtendFieldTableAndAlias(ff, fv, tag)
	//	}
	//
	//	// 当table为空时,使用第一个struct初始化table
	//	if m.table == "" {
	//		if table == "" {
	//			if table = CallTableMethod(fv); table == "" {
	//				table = m.styler.table(ff.Name)
	//			}
	//		}
	//		m.table = table
	//	}
	//
	//	// 当alias为空时,使用第一个struct初始化alias
	//	if m.alias == "" && alias != "" {
	//		m.alias = alias
	//	}
	//}
	//
	//// 如果是指针类型,还需要解指针引用
	//if ptr {
	//	fv = fv.Elem()
	//}
	//
	//// 递归下级字段,将解析的alias传递给struct中的字段
	//sub := m.parseFieldMapper(fv.Type(), fv, alias)
	//return &Field{
	//	tag: tag,
	//	extend: &ExtendField{
	//		alias:  alias,
	//		fields: &sub,
	//	},
	//}, nil
}

// 自定义类型
func (m *Mapper) structCustomTypeField(fv reflect.Value,
	addr interface{}, tag *Tag, alias string) *Field {
	column := m.decideColumnName(tag, alias)
	field := &Field{
		tag:       tag,
		addr:      addr,
		typ:       "custom",
		tagString: column,
		fv:        &fv,
	}
	m.tagFieldMapper[column] = field
	m.orderFName = append(m.orderFName, column)
	return field
}

// time类型
func (m *Mapper) structTimeTypeField(fv reflect.Value,
	addr interface{}, tag *Tag, alias string) *Field {
	column := m.decideColumnName(tag, alias)
	field := &Field{
		tag:       tag,
		addr:      addr,
		typ:       "time",
		tagString: column,
		fv:        &fv,
	}
	m.tagFieldMapper[column] = field
	m.orderFName = append(m.orderFName, column)
	return field
}

// 决定扩展字段的别名
func (m *Mapper) decideExtendFieldTableAndAlias(ff reflect.StructField, fv reflect.Value, tag *Tag) (string, string) {
	var table, alias string
	// 用户如果定义了extend相关属性,则使用定义的
	// 否则调用用户定义的方法
	// 如果方法没有定义则对字段名进行定义, 这时的alias = field name
	if tag.extendTable != "" {
		if tag.extendAlias != "" {
			table, alias = tag.extendTable, tag.extendAlias
		} else {
			// 在join map中找到对应的alias
			if m.joinMap != nil {
				if v, ok := (*m.joinMap)[tag.extendTable]; ok {
					alias = v
				}
			} else {
				alias = tag.extendTable
			}
		}
	} else {
		if table = CallTableMethod(fv); table == "" {
			// 没有为struct定义 Table方法
			// 进行名称解析
			table = m.styler.table(ff.Name)
			// 在join map中找到对应的alias
			if m.joinMap != nil {
				if v, ok := (*m.joinMap)[table]; ok {
					alias = v
				}
			} else {
				alias = table
			}
		} else {
			// 如果能从用户为 struct定义的Table中解析出alias, 则使用解析出的
			// 否则查找join map, 如果找不到, alias = table
			if table, alias = m.syntax.ParseTable(table); alias == "" {
				// 在join map中找到对应的alias
				if m.joinMap != nil {
					if v, ok := (*m.joinMap)[table]; ok {
						alias = v
					}
				} else {
					alias = table
				}
			}
		}
	}
	return table, alias
}

// 处理基本类型
func (m *Mapper) basicTypeField(f reflect.StructField, fv reflect.Value, tag *Tag, alias string) (*Field, error) {
	column := m.decideColumnName(tag, alias)
	addr, err := m.decideColumnAddr(false, fv)
	if err != nil {
		return nil, err
	}

	field := &Field{
		tag:       tag,
		typ:       f.Type.String(),
		tagString: column,
		addr:      addr,
		fv:        &fv,
	}
	m.tagFieldMapper[column] = field
	m.orderFName = append(m.orderFName, column)
	return field, nil
}

// 处理slice类型
func (m *Mapper) sliceFieldType(f reflect.StructField, fv reflect.Value, tag *Tag,
	alias string, joinMap *map[string]string) (*Field, error) {

	panic("implement me")
	//switch fv.Type().Elem().Kind() {
	//case reflect.Uint8:
	//	fmt.Println(m.processDefaultFieldType(f, fv, tag, alias))
	//}
	//return nil, nil
}

// 决定列的名称
func (m *Mapper) decideColumnName(tag *Tag, alias string) string {
	if tag.useAlias {
		if alias == "" {
			return tag.column
		}

		return alias + "." + tag.column
	}
	return tag.column
}

// 决定列的地址
func (m *Mapper) decideColumnAddr(ptr bool, fv reflect.Value) (interface{}, error) {
	// 非指针字段 获取字段的地址
	if !ptr {
		var addr interface{}
		if fv.CanAddr() {
			// 不能获取指针字段的地址
			if !fv.Addr().CanInterface() {
				return nil, errors.New("the field cannot get interface")
			}
			addr = fv.Addr().Interface()

		} else {
			// 不能获取地址
			if !fv.CanInterface() {
				return nil, errors.New("the field cannot get interface")
			}
			addr = fv.Interface()
		}
		return addr, nil
	} else {
		if !fv.CanInterface() {
			return nil, errors.New("the field cannot get interface")
		}
		return fv.Interface(), nil
	}
}

// 为映射对象中的字段地址赋值
func (m *Mapper) AssignAddressValue() {
	for _, f := range m.tagFieldMapper {
		f.assignValue()
	}
	//recursionAssignAddressValue(m.fields)
}

func recursionAssignAddressValue(fields []*Field) {
	for i, n := 0, len(fields); i < n; i++ {
		if fields[i].extend != nil {
			recursionAssignAddressValue(*fields[i].extend.fields)
		}
		fields[i].assignValue()
	}
}

func (m *Mapper) GetInsertColumnAndValues() (columns []string, values []interface{}) {
	for _, c := range m.orderFName {
		if f, ok := m.tagFieldMapper[c]; ok {
			values = append(values, f.getInsertValue())
			columns = append(columns, c)
		}
	}
	//columns, values = recursionGetColumnsAndValues(m.fields, true)
	return
}

func (m *Mapper) GetInsertValues() (values []interface{}) {
	for _, c := range m.orderFName {
		if f, ok := m.tagFieldMapper[c]; ok {
			values = append(values, f.getInsertValue())
		}
	}
	return
}

func (m *Mapper) GetValuesByColumns(columns []string) (values []interface{}) {
	if len(m.fields) <= 0 {
		return
	}
	for _, c := range columns {
		if f, ok := m.tagFieldMapper[c]; ok {
			values = append(values, f.getInsertValue())
		}
	}
	return
}

func (m *Mapper) GetColumnsAndValuesNotZero() (columns []string, values []interface{}) {
	for _, c := range m.orderFName {
		if f, ok := m.tagFieldMapper[c]; ok {
			if value := f.getUpdateValue(); value != nil {
				values = append(values, value)
				columns = append(columns, c)
			}
		}
	}
	//columns, values = recursionGetColumnsAndValues(m.fields, false)
	return
}

func recursionGetColumnsAndValues(fields []*Field, allowZero bool) (columns []string, values []interface{}) {
	var fieldsSize int
	if fieldsSize = len(fields); fieldsSize <= 0 {
		return
	}
	for i := 0; i < fieldsSize; i++ {
		if fields[i].extend != nil {
			subCols, subValues := recursionGetColumnsAndValues(*fields[i].extend.fields, allowZero)
			columns = append(columns, subCols...)
			values = append(values, subValues...)
		} else {
			// if zero not allow and the field is zero, continue
			if !allowZero && fields[i].isZero() {
				continue
			}
			columns = append(columns, fields[i].tagString)
			values = append(values, fields[i].addr)
		}
	}
	return
}
