package binding

import (
	"reflect"
)

type Binding struct {
	keysOrder []string
	args      map[string][]interface{}
}

func NewBinding() *Binding {
	return &Binding{
		args: map[string][]interface{}{
			"select": make([]interface{}, 0),
			"from":   make([]interface{}, 0),
			"join":   make([]interface{}, 0),
			"where":  make([]interface{}, 0),
			"having": make([]interface{}, 0),
			"order":  make([]interface{}, 0),
			"union":  make([]interface{}, 0),
		},
		keysOrder: []string{
			"select", "from", "join", "where", "having", "order", "union",
		},
	}
}

func (b *Binding) AddBinding(typ string, val interface{}) {
	if _, ok := b.args[typ]; ok {
		b.args[typ] = append(b.args[typ], val)
	}
}

func (b *Binding) GetBindings() (bindings []interface{}) {
	for _, key := range b.keysOrder {
		for _, v := range b.args[key] {
			mergeBindings(&bindings, v)
		}
	}
	return bindings
}

func (b *Binding) PrepareUpdateBinding(values []interface{}) (bindings []interface{}) {
	bindings = append(bindings, values...)
	keys := []string{"from", "where", "having", "order", "union"}
	for _, k := range keys {
		for _, v := range b.args[k] {
			mergeBindings(&bindings, v)
		}
	}
	return
}

func (b *Binding) PrepareDeleteBinding() (bindings []interface{}) {
	keys := []string{"where"}
	for _, k := range keys {
		for _, v := range b.args[k] {
			mergeBindings(&bindings, v)
		}
	}
	return
}

func mergeBindings(bindings *[]interface{}, v interface{}) {
	switch v.(type) {
	case
		int, int8, int32, int64,
		uint, uint8, uint32, uint64,
		float32, float64, string, bool:
		*bindings = append(*bindings, v)
	case
		[]int, []int8, []int16, []int32, []int64,
		[]uint, []uint8, []uint16, []uint32, []uint64,
		[]float32, []float64, []string, []bool:
		v := reflect.ValueOf(v)
		for i, n := 0, v.Len(); i < n; i++ {
			*bindings = append(*bindings, v.Index(i).Interface())
		}
	case []interface{}:
		*bindings = append(*bindings, v.([]interface{})...)
	}
}
