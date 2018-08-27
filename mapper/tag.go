package mapper

import (
	"reflect"
	"strings"
	"strconv"
	"github.com/Soul-Mate/sprydb/syntax"
)

const (
	spryTag         = "spry"
	columnTag       = "col"
	extendTag       = "extend"
	ignoreTag       = "ignore"
	useAliasTag     = "use_alias"
	updateZeroTag   = "update_zero"
	ignoreSymbolTag = "-"
)

type Tag struct {
	ignore      bool
	extend      bool
	column      string
	useAlias    bool
	updateZero  bool
	extendTable string
	extendAlias string
	fv          *reflect.Value
	f           *reflect.StructField
}

func NewTag(field *reflect.StructField, fieldValue *reflect.Value) *Tag {
	tag := new(Tag)
	tag.f = field
	tag.fv = fieldValue
	tag.useAlias = true
	return tag
}

func (t *Tag) parse(fieldParser func(string) string, st syntax.Syntax) *Tag {
	// 字段必须是可导出的
	if !t.fv.CanInterface() {
		t.ignore = true
		return t
	}

	tag := t.f.Tag.Get(spryTag)
	if ignore := t.isIgnore(tag); ignore {
		t.ignore = ignore
		return t
	}

	var tagAttributeVal, tagAttributeKey string
	// 解析以;号分割的多个tag属性 column extend 等
	tagGroup := strings.Split(tag, ";")
	for _, group := range tagGroup {
		// 解析每个属性的值
		tagAttributeGroup := strings.Split(group, ":")
		switch len(tagAttributeGroup) {
		case 0: // 没有属性
			continue
		case 1: // 属性值为空
			tagAttributeVal = ""
		default: // 属性键值对
			tagAttributeVal = tagAttributeGroup[1]
		}
		// 判断属性分类
		tagAttributeKey = strings.TrimSpace(tagAttributeGroup[0])
		tagAttributeVal = strings.TrimSpace(tagAttributeVal)
		switch tagAttributeKey {
		case columnTag: // col
			t.column = tagAttributeVal
		case useAliasTag: // alias. 如果设置了该属性为false, 则对field解析的column不增加alias
			if v, err := strconv.ParseBool(tagAttributeVal); err == nil {
				t.useAlias = v
			}
		case updateZeroTag: // 控制结构体的字段零值是否更新
			if v, err := strconv.ParseBool(tagAttributeVal); err == nil {
				t.updateZero = v
			}
		case extendTag:
			t.extend = true
			t.extendTable, t.extendAlias = st.ParseTable(tagAttributeVal)
		}
	}

	if t.column == "" {
		t.processDefaultColumn(fieldParser)
	}

	return t
}

// 忽略字段
func (t *Tag) isIgnore(tag string) bool {
	if tag == "" {
		switch t.f.Type.Kind() {
		case reflect.Ptr: // 指针类型不是struct的都忽略
			if t.f.Type.Elem().Kind() == reflect.Struct {
				return false
			}
			return true
		case reflect.Struct:
			return false
		default:
			return true
		}
	}

	if tag == ignoreSymbolTag || tag == ignoreTag {
		return true
	}

	return false
}

func (t *Tag) processDefaultColumn(fieldParser func(string) string) {
	if column := recursionKind(*t.f, t.f.Type, fieldParser); column != "" {
		t.column = column
	}
}

func recursionKind(f reflect.StructField, ft reflect.Type, fieldParser func(string) string) (column string) {
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
