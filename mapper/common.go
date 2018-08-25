package mapper

import "reflect"

func CallTableMethod(v reflect.Value) string {
	method := v.MethodByName("Table")
	if !method.IsValid() {
		return ""
	}
	ret := method.Call(nil)[0]
	if ret.Kind() != reflect.String {
		return ""
	}
	return ret.String()
}

func CallPKMethod(v reflect.Value) string {
	method := v.MethodByName("PK")
	if !method.IsValid() {
		return "id"
	}
	ret := method.Call(nil)[0]
	if ret.Kind() != reflect.String {
		return "id"
	}
	return ret.String()
}
