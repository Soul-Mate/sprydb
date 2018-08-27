package mapper

import (
	"testing"
	"reflect"
)

type TestTagObject struct {
	Id int `spry:"column:id"`
	Name string `spry:"column:name"`
	Ignore string `spry:"ignore"`
	IgnoreSymbol string `spry:"-"`
	id int
	S1 struct{}

}

func TestNewTag(t *testing.T) {
	style := &UnderlineMapperStyle{}
	objValue := reflect.ValueOf(TestTagObject{})
	objType := objValue.Type()
	for i := 0; i< objType.NumField(); i++ {
		pff := objType.Field(i)
		pfv := objValue.Field(i)
		tag := newTag(&pff, &pfv)
		tag.parse(style.column)
	}
}