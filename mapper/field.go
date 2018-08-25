package mapper

import (
	"reflect"
	"database/sql"
	"time"
)

type ExternalField struct {
	alias  string
	fields *[]*Field
}

// if field type is Custom, the field holder value is rawData
type Field struct {
	typ         string
	tagString   string
	addr        interface{}
	v           *reflect.Value
	tag         *Tag
	rawData     []byte
	nullInt64   sql.NullInt64
	nullFloat64 sql.NullFloat64
	nullString  sql.NullString
	nullBool    sql.NullBool
	external    *ExternalField
}

// get sql.Null<T> type pointer
func (f *Field) getSqlNullType() interface{} {
	switch f.typ {
	case
		"int", "uint",
		"int8", "uint8",
		"int16", "uint16",
		"int32", "uint32",
		"int64", "uint64":
		return &f.nullInt64
	case "float32", "float64":
		return &f.nullFloat64
	case "string":
		return &f.nullString
	case "bool":
		return &f.nullBool
	case "custom", "time":
		return &f.rawData
	default:
		return &f.addr
	}
}

// get the value from the previous sql.Null<T> and assign it to the address stored in the field
func (f *Field) assignValue() {
	switch f.typ {
	case "int":
		*f.addr.(*int) = int(f.nullInt64.Int64)
	case "int8":
		*f.addr.(*int8) = int8(f.nullInt64.Int64)
	case "int16":
		*f.addr.(*int16) = int16(f.nullInt64.Int64)
	case "int32":
		*f.addr.(*int32) = int32(f.nullInt64.Int64)
	case "int64":
		*f.addr.(*int64) = f.nullInt64.Int64
	case "uint":
		*f.addr.(*uint) = uint(f.nullInt64.Int64)
	case "uint8":
		*f.addr.(*uint8) = uint8(f.nullInt64.Int64)
	case "uint16":
		*f.addr.(*uint16) = uint16(f.nullInt64.Int64)
	case "uint32":
		*f.addr.(*uint32) = uint32(f.nullInt64.Int64)
	case "uint64":
		*f.addr.(*uint64) = uint64(f.nullInt64.Int64)
	case "float32":
		*f.addr.(*float32) = float32(f.nullFloat64.Float64)
	case "float64":
		*f.addr.(*float64) = f.nullFloat64.Float64
	case "string":
		*f.addr.(*string) = f.nullString.String
	case "bool":
		*f.addr.(*bool) = f.nullBool.Bool
	case "time":
		var layout string
		if tm, ok := (*f).addr.(Time); ok {
			// 使用默认模板
			if tm.layout == "" {
				layout = "2006-01-02 15:04:05"
			}
			if t, err := time.Parse(layout, string(f.rawData)); err == nil {
				tm.Time = t
			}
		} else if tm, ok := (*f).addr.(*Time); ok {
			if tm.layout == "" {
				layout = "2006-01-02 15:04:05"
			}
			if t, err := time.Parse(layout, string(f.rawData)); err == nil {
				tm.Time = t
			}
		}
	case "custom":
		(*f).addr.(Custom).Read((*f).rawData)
	default:
	}
}

func (f *Field) getUpdateValue() interface{} {
	switch f.typ {
	case "time":
		if f.isZero() && !f.tag.updateZero {
			return nil
		}
		if tm, ok := (*f).addr.(Time); ok {
			return tm.Format(tm.layout)
		}
		if tm, ok := (*f).addr.(*Time); ok {
			return tm.Format(tm.layout)
		}
		return nil
	case "custom":
		if data := (*f).addr.(Custom).Write(); len(data) <= 0 && !f.tag.updateZero {
			return nil
		} else {
			return data
		}
	default:
		if !f.tag.updateZero && f.isZero() {
			return nil
		}
		return (*f).addr
	}
}

func (f *Field) getInsertValue() interface{} {
	switch f.typ {
	case "time":
		if f.isZero() {
			return nil
		}
		if tm, ok := (*f).addr.(Time); ok {
			return tm.Format(tm.layout)
		}
		if tm, ok := (*f).addr.(*Time); ok {
			return tm.Format(tm.layout)
		}
		return nil
	case "custom":
		return (*f).addr.(Custom).Write()
	default:
		return (*f).addr
	}
}

func (f *Field) getFieldValue() interface{} {
	switch f.typ {
	case "time":
		if f.isZero() {
			return nil
		}
		if tm, ok := (*f).addr.(Time); ok {
			return tm.Format(tm.layout)
		}
		if tm, ok := (*f).addr.(*Time); ok {
			return tm.Format(tm.layout)
		}
		return nil
	case "custom":
		if data := (*f).addr.(Custom).Write(); len(data) <= 0 {
			return nil
		} else {
			return data
		}
	default:
		return (*f).addr
	}
}

func (f *Field) isZero() bool {
	switch f.v.Kind() {
	case reflect.String:
		return f.v.Len() == 0
	case reflect.Bool:
		return !f.v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return f.v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return f.v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return f.v.IsNil()
	}
	return reflect.DeepEqual(f.v.Interface(), reflect.Zero(f.v.Type()).Interface())
}