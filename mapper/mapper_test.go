package mapper

import (
	"testing"
	"errors"
	"github.com/Soul-Mate/sprydb/syntax"
	)

var syntax2 = syntax.NewSyntax("mysql")

type CustomText struct {
	text string
	len  int
}

func (c *CustomText) Read(data []byte) {
	c.text = string(data)
	c.len = len(c.text)
}

func (c *CustomText) Write() ([]byte) {
	return nil
}

func TestMapper_ParseField(t *testing.T) {
	var err error
	if err = parseFieldDefineColumn(); err != nil {
		t.Error(err)
	}
	if err = parseFieldIgnoreField(); err != nil {
		t.Error(err)
	}
	if err = parseFiledPointerField(); err != nil {
		t.Error(err)
	}
	if err = parseFieldSubField(); err != nil {
		t.Error(err)
	}
	if err = parseFieldExternalField(); err != nil {
		t.Error(err)
	}
	if err = parseCustomField(); err != nil {
		t.Error(err)
	}
}

func parseFieldDefineColumn() error {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       int
		Name     string `spry:"column:base_name"`
		Password string `spry:"column:sha1_password"`
		Token    string `spry:"column:access_token"`
		Sub      struct {
			SubId       int    `spry:"column:subId"`
			SubName     string `spry:"column:subName"`
			SubPassword string `spry:"column:subPassword"`
			SubToken    string `spry:"column:subToken"`
		}
	}{}
	parseError := errors.New("parseFieldDefineColumn: parse error")
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		return err
	}
	if err = objMapper.Parse(); err != nil {
		return err
	}
	columns, address := objMapper.GetColumnAndAddress()
	for _, col := range columns {
		switch col {
		case "id":
		case "base_name":
		case "sha1_password":
		case "access_token":
		case "subId":
		case "subName":
		case "subPassword":
		case "subToken":
		default:
			return parseError
		}
	}
	if len(address) != 8 {
		return parseError
	}
	return nil
}

func parseFieldIgnoreField() error {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       int
		Name     string `spry:"ignore"`
		Password string
		Token    string
		Sub      struct {
			SubId       int `spry:"ignore"`
			SubName     string
			SubPassword string
			SubToken    string
		}
	}{}
	parseError := errors.New("parseFieldIgnoreField: parse error")
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		return err
	}

	if err = objMapper.Parse(); err != nil {
		return err
	}
	columns, address := objMapper.GetColumnAndAddress()
	for _, col := range columns {
		switch col {
		case "id":
		case "name":
			return parseError
		case "password":
		case "token":
		case "sub_id":
			return parseError
		case "sub_name":
		case "sub_password":
		case "sub_token":
		default:
			return parseError
		}
	}
	if len(address) != 6 {
		return parseError
	}
	return nil
}

func parseFiledPointerField() error {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       *int
		Name     *string
		Password *string
		Token    *string
	}{}
	parseError := errors.New("parseFiledPointerField: parse error")
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		return err
	}
	if err = objMapper.Parse(); err != nil {
		return err
	}
	columns, address := objMapper.GetColumnAndAddress()
	for _, col := range columns {
		switch col {
		case "id":
		case "name":
		case "password":
		case "token":
		default:
			return parseError
		}
	}
	if len(columns) != 4 || len(address) != 4 {
		return parseError
	}
	return nil
}

func parseFieldSubField() error {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       int
		Name     string
		Password string
		Token    string
		Sub      struct {
			SubId       int
			SubName     string
			SubPassword string
			SubToken    string
		}
	}{}
	parseError := errors.New("parseFieldSubField: parse error")
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		return err
	}
	objMapper.SetAlias("")
	objMapper.SetJoinMap(nil)
	if err = objMapper.Parse(); err != nil {
		return err
	}
	columns, address := objMapper.GetColumnAndAddress()
	for _, col := range columns {
		switch col {
		case "id":
		case "name":
		case "password":
		case "token":
		case "sub_id":
		case "sub_name":
		case "sub_password":
		case "sub_token":
		default:
			return parseError
		}
	}
	if len(address) != 8 {
		return parseError
	}
	return nil
}

func parseFieldExternalField() error {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       int
		Name     string
		Password string
		Token    string
		Sub      *struct {
			SubId       int
			SubName     string
			SubPassword string
			SubToken    string `spry:"ignore"`
			SubSub      *struct {
				Id   int `spry:"column:id"`
				Name int `spry:"column:name"`
			} `spry:"extend:sub_sub"`
		} `spry:"extend"`
	}{}
	parseError := errors.New("parseFieldExternalField: parse error")
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		return err
	}
	if err = objMapper.Parse(); err != nil {
		return err
	}
	columns, address := objMapper.GetColumnAndAddress()
	for _, col := range columns {
		switch col {
		case "id":
		case "name":
		case "password":
		case "token":
		case "sub.sub_id":
		case "sub.sub_name":
		case "sub.sub_password":
		case "sub_sub.id":
		case "sub_sub.name":
		default:
			return parseError
		}
	}
	if len(address) != 9 {
		return parseError
	}
	return nil
}

func parseCustomField() error {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       int
		Name     string
		Password string
		Token    string
		Sub      *struct {
			SubId       int
			SubName     string
			SubPassword string
			SubToken    string `spry:"ignore"`
			SubSub      *struct {
				Id   int `spry:"column:id"`
				Name int `spry:"column:name"`
			} `spry:"extend:sub_sub"`
		} `spry:"extend"`
		CustomText
	}{}
	parseError := errors.New("parseCustomField: parse error")
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		return err
	}
	objMapper.SetAlias("")
	objMapper.SetJoinMap(nil)
	if err = objMapper.Parse(); err != nil {
		return err
	}
	columns, address := objMapper.GetColumnAndAddress()
	for _, col := range columns {
		switch col {
		case "id":
		case "name":
		case "password":
		case "token":
		case "sub.sub_id":
		case "sub.sub_name":
		case "sub.sub_password":
		case "sub_sub.id":
		case "sub_sub.name":
		case "custom_text":
		default:
			return parseError
		}
	}
	if len(address) != 10 {
		return parseError
	}
	return nil
}

func TestMapper_GetPK(t *testing.T) {
	type obj struct {}
	mapper, err := NewMapper(&obj{}, syntax2, nil)
	if err != nil {
		t.Error(err)
	}
	if mapper.GetPK() != "id" {
		t.Error("TestMapper_GetPK error")
	}
}

func TestMapper_GetTable(t *testing.T) {
	type obj struct {}
	mapper, err := NewMapper(&obj{}, syntax2, nil)
	if err != nil {
		t.Error(err)
	}
	if mapper.GetTable() != "obj" {
		t.Error("TestMapper_GetPK error")
	}
	// set table
	mapper, err = NewMapper(&obj{}, syntax2, nil)
	if err != nil {
		t.Error(err)
	}
	mapper.SetTable("object")
	if mapper.GetTable() != "object" {
		t.Error("TestMapper_GetPK error")
	}


}

func TestMapper_GetAlias(t *testing.T) {
	type obj struct {}
	// set alias
	mapper, err := NewMapper(&obj{}, syntax2, nil)
	if err != nil {
		t.Error(err)
	}
	mapper.SetAlias("a")
	if mapper.GetTable() != "obj" {
		t.Error("TestMapper_GetPK error")
	}
	if mapper.GetAlias() != "a" {
		t.Error("TestMapper_GetPK error")
	}
}

func TestMapper_GetColumn(t *testing.T) {
	var (
		err       error
		objMapper *Mapper
	)
	obj := struct {
		Id       int
		Name     string
		Password string
		Token    string
		Sub      *struct {
			SubId       int
			SubName     string
			SubPassword string
			SubToken    string `spry:"ignore"`
			SubSub      *struct {
				Id   int `spry:"column:sub_sub_id"`
				Name int `spry:"column:name"`
			} `spry:"extend:sub_sub"`
		} `spry:"extend:sub"`
		CustomText
	}{}
	if objMapper, err = NewMapper(&obj, syntax2, nil); err != nil {
		t.Error(err)
	}
	objMapper.SetTable("foo")
	if err = objMapper.Parse(); err != nil {
		t.Error(err)
	}
	rawColumn := []string{
		"id", "name", "password", "token",
		"sub.sub_id", "sub.sub_name", "sub.sub_password",
		"sub_sub.sub_sub_id", "sub_sub.name",
	}
	column := objMapper.GetColumn()
	for k, v := range rawColumn {
		if column[k] != v {
			t.Error("TestMapper_GetColumn error")
		}
	}

	if len(column) != 10 {
		t.Error("TestMapper_GetColumn error")
	}
}

