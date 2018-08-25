package mapper

import (
	"reflect"
	"strings"
	"strconv"
)

const (
	SqlTag      = "spry"
	ColumnTag   = "col"
	UseAliasTag = "alias"
	ExternalTag = "external"
	UpdateZero  = "update_zero"
)

type Tag struct {
	ignore        bool
	isExt         bool
	column        string
	useAlias      bool
	updateZero    bool
	externalAlias string
	fv            *reflect.Value
	f             *reflect.StructField
}

func NewTag(field *reflect.StructField, fieldValue *reflect.Value) *Tag {
	tag := new(Tag)
	tag.f = field
	tag.fv = fieldValue
	tag.useAlias = true
	return tag
}

func (t *Tag) parse(fieldParser func(string) string) *Tag {
	// field must be export
	if !t.fv.CanInterface() {
		t.ignore = true
		return t
	}
	tag := t.f.Tag.Get(SqlTag)
	if tag == "-" || tag == "ignore" {
		t.ignore = true
		return t
	}
	// if there is no tag defined, will be resolved by field type
	if tag == "" {
		t.processDefaultColumn(fieldParser)
	} else {
		var tagVal, tagKey string
		// 解析以分好分割的多个tag属性 column external 等
		tagGroup := strings.Split(tag, ";")
		for _, group := range tagGroup {
			// 解析每个属性的值
			valueGroup := strings.Split(group, ":")
			switch len(valueGroup) {
			case 0: // 没有属性
				continue
			case 1: // 属性值为空
				tagVal = ""
			default: // 属性键值对
				tagVal = valueGroup[1]
			}
			// 判断属性分类
			tagKey = strings.TrimSpace(valueGroup[0])
			tagVal = strings.TrimSpace(tagVal)
			switch tagKey {
			case ColumnTag: // 定义的column
				t.column = tagVal
			case UseAliasTag: // 忽略alias 如果设置了该属性为false,则对field解析的column不增加alias
				if v, err := strconv.ParseBool(tagVal); err == nil {
					t.useAlias = v
				}
			case UpdateZero: // 控制结构体的字段零值是否更新
				if v, err := strconv.ParseBool(tagVal); err == nil {
					t.updateZero = v
				}
			case ExternalTag:
				t.isExt = true
				t.externalAlias = tagVal
			}
		}
		if t.column == "" {
			t.processDefaultColumn(fieldParser)
		}
	}
	return t
}

func (t *Tag) processDefaultColumn(fieldParser func(string) string) {
	if column := recursionKind(*t.f, t.f.Type, fieldParser); column != "" {
		t.column = column
	}
}

func recursionKind(f reflect.StructField, ft reflect.Type,
	fieldParser func(string) string) (column string) {
	switch ft.Kind() {
	//case reflect.Struct:
	//	if reflect.PtrTo(ft).Implements(reflect.TypeOf((*Custom)(nil)).Elem()) {
	//		column = fieldParser(f.Name)
	//	}
	//case reflect.Map, reflect.Complex64, reflect.Complex128, reflect.Func, reflect.Chan:
	//	return ""
	case reflect.Ptr:
		column = recursionKind(f, f.Type.Elem(), fieldParser)
	default:
		column = fieldParser(f.Name)
	}
	return column
}
