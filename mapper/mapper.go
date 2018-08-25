package mapper

import (
	"reflect"
	"github.com/Soul-Mate/sprydb/define"
	"fmt"
	"errors"
	"regexp"
	"github.com/Soul-Mate/sprydb/syntax"
	"time"
)

type MapperInterface interface {
	ParseField() error
	GetTable() (table string)
	GetPK() (pk string)
	SetTable(table string)
	SetTableAlias(alias string)
	SetAliasMap(aliasMap *map[string]string)
	GetColumn() (column []string)
	GetAddress() (address []interface{})
	GetAddressByColumn(column []string) (address []interface{})
	GetColumnAndAddress() (column []string, address []interface{})
	GetValuesByColumns(columns []string) ([]interface{})
	GetColumnsAndValues() (columns []string, values []interface{})
}

type Mapper struct {
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
		m.pk = CallPKMethod(m.v)
	}
	return m.pk
}

func (m *Mapper) GetTable() string {
	return m.table
	//if m.table == "" {
	//	// table还是为空,则进行名称解析
	//	if m.table = CallTableMethod(m.v); m.table == "" {
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
		if (*fs)[i].external != nil {
			subColumn, subAddress, err := recursionGetColumnAndAddress((*fs)[i].external.fields)
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

func (m *Mapper) Parse() error {
	var fv reflect.Value
	if m.ptr {
		fv = m.pv
	} else {
		fv = m.v
	}
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
		if tag = NewTag(&ff, &fv).parse(m.styler.column); tag.ignore {
			continue
		}
		// 处理字段类型
		if field, err := m.processFieldType(ff, fv, tag, alias); err == nil {
			fields = append(fields, field)
		}
	}
	return
}

// 处理字段类型
func (m *Mapper) processFieldType(ff reflect.StructField, fv reflect.Value,
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
		return m.ptrStructFieldType(ff, fv, tag, alias)
	default:
		return m.baseTypeField(ff, fv, tag, alias), nil
	}
}

func (m *Mapper) ptrStructFieldType(ff reflect.StructField, pfv reflect.Value,
	tag *Tag, alias string) (*Field, error) {
	// 如果是一个空指针
	if !pfv.Elem().IsValid() {
		// 如果该指针不能设置内存则跳过这个字段
		if !pfv.CanSet() {
			return nil, errors.New(
				fmt.Sprintf("The %s field is a pointer type, but the address cannot be set, you can set ignore",
					ff.Type.Elem().Kind().String()))
		}
		// 对空指针设置内存
		pfv.Set(reflect.New(ff.Type.Elem()))
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
		return m.baseTypeField(ff, elem, tag, m.alias), nil
	}
}

// 处理类型是struct的字段以及特殊字段
// 如果ptr是true, ff 和 fv是指针类型
func (m *Mapper) structTypeField(ff reflect.StructField, fv reflect.Value,
	tag *Tag, ptr bool, alias string) (*Field, error) {
	var (
		addr  interface{}
		table string
	)

	// 非指针字段 获取字段的地址
	if !ptr {
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
	} else {
		if !fv.CanInterface() {
			return nil, errors.New("the field cannot get interface")
		}
		addr = fv.Interface()
	}

	// 特殊类型字段进行特殊处理
	switch addr.(type) {
	case time.Time, *time.Time: // time字段
		return m.structTimeTypeField(fv, addr, tag, alias), nil
	case Custom: // 用户自定义字段
		return m.structCustomTypeField(fv, addr, tag, alias), nil
	default:
		// 如果是join的struct
		// 定义了alias,则用这个alias覆盖上级的alias
		// 否则会调用定义Table()进行解析
		// 否则统一使用上级的alias
		if tag.isExt && tag.useAlias {
			if tag.externalAlias != "" {
				alias = tag.externalAlias
			} else {
				table, alias = m.parseStructTableAndAlias(ff, fv, tag)
			}

			// 当table为空时,使用第一个struct初始化table
			if m.table == "" {
				if table == "" {
					if table = CallTableMethod(fv); table == "" {
						table = m.styler.table(ff.Name)
					}
				}
				m.table = table
			}

			// 当alias为空时,使用第一个struct初始化alias
			if m.alias == "" && alias != "" {
				m.alias = alias
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
			external: &ExternalField{
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
	//		v:         &fv,
	//	}
	//	m.tagFieldMapper[column] = field
	//	m.orderFName = append(m.orderFName, column)
	//	return field, nil
	//}

	// 如果是join的struct
	// 定义了alias,则用这个alias覆盖上级的alias
	// 否则会调用定义Table()进行解析
	// 否则统一使用上级的alias
	//if tag.isExt && tag.useAlias {
	//	if tag.externalAlias != "" {
	//		alias = tag.externalAlias
	//	} else {
	//		table, alias = m.parseStructTableAndAlias(ff, fv, tag)
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
	//	external: &ExternalField{
	//		alias:  alias,
	//		fields: &sub,
	//	},
	//}, nil
}

// 自定义类型
func (m *Mapper) structCustomTypeField(fv reflect.Value,
	addr interface{}, tag *Tag, alias string) *Field {
	var column string
	if !tag.useAlias {
		column = tag.column
	} else {
		if alias != "" {
			column = alias + "." + tag.column
		} else {
			column = tag.column
		}
	}

	field := &Field{
		tag:       tag,
		addr:      addr,
		typ:       "custom",
		tagString: column,
		v:         &fv,
	}
	m.tagFieldMapper[column] = field
	m.orderFName = append(m.orderFName, column)
	return field
}

// time类型
func (m *Mapper) structTimeTypeField(fv reflect.Value,
	addr interface{}, tag *Tag, alias string) *Field {
	var column string
	if !tag.useAlias {
		column = tag.column
	} else {
		if alias != "" {
			column = alias + "." + tag.column
		} else {
			column = tag.column
		}
	}

	field := &Field{
		tag:       tag,
		addr:      addr,
		typ:       "time",
		tagString: column,
		v:         &fv,
	}
	m.tagFieldMapper[column] = field
	m.orderFName = append(m.orderFName, column)
	return field
}

// 解析连接的struct的table和alias
func (m *Mapper) parseStructTableAndAlias(ff reflect.StructField, fv reflect.Value, tag *Tag) (table, alias string) {
	// 调用用户定义的Table()方法
	if table = CallTableMethod(fv); table == "" {
		// 调用失败解析名称
		table = m.styler.table(ff.Name)
		alias = table
		return
	}
	// 在Join map中查找对应的alias
	// 如果查找到则直接返回
	if m.joinMap != nil {
		if v, ok := (*m.joinMap)[table]; ok {
			alias = v
			return
		}
	}
	// 调用Table()方法后的返回值进行解析alias 类似 foo as bar
	// 这样的格式可以解析出table为foo alias 为bar
	if table, alias = m.syntax.ParseTable(table); alias == "" {
		// 如果解析后alias为空
		// 则在joinMap中查找, 查找到则返回
		if m.joinMap != nil {
			if v, ok := (*m.joinMap)[table]; ok {
				alias = v
				return
			}
		}
		// 没有查找到则alias == table
		alias = table
		return
	}
	return
}

// 处理基本类型
func (m *Mapper) baseTypeField(f reflect.StructField, fv reflect.Value, tag *Tag, alias string) *Field {
	var (
		column string
		iface  interface{}
	)
	if tag.useAlias {
		if alias != "" {
			column = alias + "." + tag.column
		} else {
			column = tag.column
		}
	} else {
		column = tag.column
	}

	if !fv.CanInterface() {
		return nil
	}

	if fv.CanAddr() {
		iface = fv.Addr().Interface()
	} else {
		iface = fv.Interface()
	}

	field := &Field{
		tag:       tag,
		typ:       f.Type.String(),
		tagString: column,
		addr:      iface,
		v:         &fv,
	}
	m.tagFieldMapper[column] = field
	m.orderFName = append(m.orderFName, column)
	return field
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

// 为映射对象中的字段地址赋值
func (m *Mapper) AssignAddressValue() {
	for _, f := range m.tagFieldMapper {
		f.assignValue()
	}
	//recursionAssignAddressValue(m.fields)
}

func recursionAssignAddressValue(fields []*Field) {
	for i, n := 0, len(fields); i < n; i++ {
		if fields[i].external != nil {
			recursionAssignAddressValue(*fields[i].external.fields)
		}
		fields[i].assignValue()
	}
}

func (m *Mapper) GetColumnsAndValues() (columns []string, values []interface{}) {
	for _, c := range m.orderFName {
		if f, ok := m.tagFieldMapper[c]; ok {
			values = append(values, f.getInsertValue())
			columns = append(columns, c)
		}
	}
	//columns, values = recursionGetColumnsAndValues(m.fields, true)
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
		if fields[i].external != nil {
			subCols, subValues := recursionGetColumnsAndValues(*fields[i].external.fields, allowZero)
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
