package query

import (
	"testing"
	"reflect"
	"github.com/Soul-Mate/sprydb/syntax"
	"github.com/Soul-Mate/sprydb/binding"
	"fmt"
)

func TestBuilder_Table(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	b := NewBuilder("mysql", syntax2, binding2)
	b.Table("foo")
	if b.tableName != "foo" || b.tableAlias != "" {
		t.Error("TestBuilder_Table error")
	}
}

func TestBuilder_Distinct(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	b := NewBuilder("mysql", syntax2, binding2)
	b.Distinct()
	if b.distinct != true {
		t.Error("TestBuilder_Distinct error")
	}
}

func TestBuilder_Select(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	b := NewBuilder("mysql", syntax2, binding2)
	columns := []string{"id", "name", "example"}
	b.Select(columns...)

	if b.column[0] != "id" {
		t.Error("TestBuilder_Select error")
	}
	if b.column[1] != "name" {
		t.Error("TestBuilder_Select error")
	}
	if b.column[2] != "example" {
		t.Error("TestBuilder_Select error")
	}
}

func TestBuilder_Where(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	grammar2 := NewGrammarFactory("mysql", syntax2, binding2)
	rawSQL := "select * from `users` where `id` > ? and `name` = ?"
	b := NewBuilder("mysql", syntax2, binding2)
	b.Table("users").Where("id", ">", 1).
		Where("name", "=", "2")
	if b.wheres[0]["column"] != "id" {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[0]["operator"] != ">" {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[0]["value"] != 1 {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[0]["logic"] != "and" {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[1]["column"] != "name" {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[1]["operator"] != "=" {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[1]["value"] != "2" {
		t.Error("TestBuilder_Where error")
	}
	if b.wheres[1]["logic"] != "and" {
		t.Error("TestBuilder_Where error")
	}
	if !reflect.DeepEqual(b.binding.GetBindings(), []interface{}{1, "2"}) {
		t.Error("TestBuilder_Where error")
	}
	buildSQL, err := grammar2.CompileSelect(b)
	if err != nil {
		t.Error(err)
	}
	if buildSQL != rawSQL {
		t.Error("TestBuilder_Where error")
	}
}

func TestBuilder_OrWhere(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	grammar2 := NewGrammarFactory("mysql", syntax2, binding2)
	rawSQL := "select * from `users` where `id` > ? or `name` = ?"
	b := NewBuilder("mysql", syntax2, binding2)
	b.Table("users").Where("id", ">", 1).
		OrWhere("name", "=", "2")
	if b.wheres[0]["column"] != "id" {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[0]["operator"] != ">" {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[0]["value"] != 1 {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[0]["logic"] != "and" {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[1]["column"] != "name" {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[1]["operator"] != "=" {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[1]["value"] != "2" {
		t.Error("TestBuilder_OrWhere error")
	}
	if b.wheres[1]["logic"] != "or" {
		t.Error("TestBuilder_OrWhere error")
	}
	if !reflect.DeepEqual(b.binding.GetBindings(), []interface{}{1, "2"}) {
		t.Error("TestBuilder_OrWhere error")
	}
	buildSQL, err := grammar2.CompileSelect(b)
	if err != nil {
		t.Error(err)
	}
	if buildSQL != rawSQL {
		t.Error("TestBuilder_Where error")
	}
}

func TestBuilder_WhereIn(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	grammar2 := NewGrammarFactory("mysql", syntax2, binding2)
	rawSQL := "select * from `users` where in (?,?,?,?,?) and in (?,?,?)"
	b := NewBuilder("mysql", syntax2, binding2)
	b.Table("users").WhereIn("id", 1, 2, 3, 4, 5).
		WhereIn("name", "a", "b", "c")
	if b.wheres[0]["column"] != "id" {
	}
	if b.wheres[0]["type"] != "In" {
		t.Error("TestBuilder_WhereIn error")
	}
	if !reflect.DeepEqual(b.wheres[0]["value"], []interface{}{1, 2, 3, 4, 5}) {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[0]["logic"] != "and" {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[1]["column"] != "name" {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[1]["type"] != "In" {
		t.Error("TestBuilder_WhereIn error")
	}
	if !reflect.DeepEqual(b.wheres[1]["value"], []interface{}{"a", "b", "c"}) {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[1]["logic"] != "and" {
		t.Error("TestBuilder_WhereIn error")
	}
	if !reflect.DeepEqual(b.binding.GetBindings(), []interface{}{1, 2, 3, 4, 5, "a", "b", "c"}) {
		t.Error("TestBuilder_WhereIn error")
	}
	buildSQL, err := grammar2.CompileSelect(b)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(buildSQL)
	fmt.Println(rawSQL)
}

func TestBuilder_OrWhereIn(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	b := NewBuilder("mysql", syntax2, binding2)
	b.OrWhereIn("id", 1, 2, 3, 4, 5).
		OrWhereIn("name", "a", "b", "c")
	if b.wheres[0]["column"] != "id" {
	}
	if b.wheres[0]["type"] != "In" {
		t.Error("TestBuilder_WhereIn error")
	}
	if !reflect.DeepEqual(b.wheres[0]["value"], []interface{}{1, 2, 3, 4, 5}) {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[0]["logic"] != "or" {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[1]["column"] != "name" {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[1]["type"] != "In" {
		t.Error("TestBuilder_WhereIn error")
	}
	if !reflect.DeepEqual(b.wheres[1]["value"], []interface{}{"a", "b", "c"}) {
		t.Error("TestBuilder_WhereIn error")
	}
	if b.wheres[1]["logic"] != "or" {
		t.Error("TestBuilder_WhereIn error")
	}
}

func TestBuilder_OrderBy(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	b := NewBuilder("mysql", syntax2, binding2)
	b.OrderBy("id", "desc")
	if b.orders["direction"] != "desc" {
		t.Error("TestBuilder_OrderBy error")
	}
	if !reflect.DeepEqual(b.orders["column"], []string{"id"}) {
		t.Error("TestBuilder_OrderBy error")
	}
	b.OrderBy("name", "DESC")
	if b.orders["direction"] != "desc" {
		t.Error("TestBuilder_OrderBy error")
	}
	if !reflect.DeepEqual(b.orders["column"], []string{"id", "name"}) {
		t.Error("TestBuilder_OrderBy error")
	}
	b.OrderBy("name", "")
	if b.orders["direction"] != "asc" {
		t.Error("TestBuilder_OrderBy error")
	}
	if !reflect.DeepEqual(b.orders["column"], []string{"id", "name", "name"}) {
		t.Error("TestBuilder_OrderBy error")
	}
}

func TestBuilder_OrderByMulti(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	b := NewBuilder("mysql", syntax2, binding2)
	b.OrderByMulti([]string{"id", "name"}, "")
	if b.orders["direction"] != "asc" {
		t.Error("TestBuilder_OrderByMulti error")
	}
	if !reflect.DeepEqual(b.orders["column"], []string{"id", "name"}) {
		t.Error("TestBuilder_OrderByMulti error")
	}
}

func TestBuilder_Join(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	grammar2 := NewGrammarFactory("mysql", syntax2, binding2)
	rawSQL := "select * from `users` inner join `levels` on `users`.`id` = `levels`.`user_id`"
	b := NewBuilder("mysql", syntax2, binding2)
	b.Table("users").Join("levels", "users.id", "=", "levels.user_id")
	buildeSQL, err := grammar2.CompileSelect(b)
	if err != nil {
		t.Error(err)
	}
	if buildeSQL != rawSQL {
		t.Error("TestBuilder_Join error")
	}
}

func TestBuilder_JoinClosure(t *testing.T) {
	syntax2  := syntax.NewSyntax("mysql")
	binding2 := binding.NewBinding()
	grammar2 := NewGrammarFactory("mysql", syntax2, binding2)
	rawSQL := "select * from `users` inner join `levels` on `levels`.`user_id` = `users`.`id` " +
		"or `levels`.`user_name` = `users`.`name` " +
		"and `levels`.`user_id` = `users`.`id` and id > ? " +
		"inner join `user_levels` on `user_levels`.`level_id` = `levels`.`id` " +
		"and user_levles.level_id > ?"
	b := NewBuilder("mysql", syntax2, binding2)
	b.Table("users").JoinClosure("levels", func(join *BuilderJoin) {
		join.On("levels.user_id", "=", "users.id").
			OrOn("levels.user_name", "=", "users.name").
			On("levels.user_id", "=", "users.id").
			Where("id", ">", 2)
	}).JoinClosure("user_levels", func(join *BuilderJoin) {
		join.On("user_levels.level_id", "=", "levels.id").
			Where("user_levles.level_id", ">", 10)
	})
	buildeSQL, err := grammar2.CompileSelect(b)
	if err != nil {
		t.Error(err)
	}
	if buildeSQL != rawSQL {
		t.Error("TestBuilder_JoinClosure error")
	}
}
