package query

import (
	"reflect"
	"fmt"
	"errors"
)

func (b *Builder) Where(column, operator string, parameters interface{}) *Builder {
	if err := b.where(column, operator, parameters, "and"); err != nil {
		b.err = err
		return b
	}
	return b
}

func (b *Builder) OrWhere(column, operator string, parameters interface{}) *Builder {
	if err := b.where(column, operator, parameters, "or"); err != nil {
		b.err = err
		return b
	}
	return b
}

func (b *Builder) where(column, operator string, parameters interface{}, logic string) (err error) {
	if operator, err = b.syntax.PrepareWhereOperator(operator); err != nil {
		return
	}
	b.wheres = append(b.wheres, map[string]interface{}{
		"type":     "Basic",
		"column":   column,
		"operator": operator,
		"value":    parameters,
		"logic":    logic,
	})
	b.binding.AddBinding("where", parameters)
	return
}

func (b *Builder) WhereIn(column string, parameters ...interface{}) *Builder {
	b.whereIn(column, "and", false, parameters...)
	return b
}

func (b *Builder) WhereNotIn(column string, parameters ...interface{}) *Builder {
	b.whereIn(column, "and", true, parameters...)
	return b
}

func (b *Builder) OrWhereIn(column string, parameters ...interface{}) *Builder {
	b.whereIn(column, "or", false, parameters...)
	return b
}

func (b *Builder) OrWhereNotIn(column string, parameters ...interface{}) *Builder {
	b.whereIn(column, "or", true, parameters...)
	return b
}

func (b *Builder) whereIn(column string, logic string, not bool, parameters ...interface{}) {
	b.wheres = append(b.wheres, map[string]interface{}{
		"type":   "In",
		"column": column,
		"value":  parameters,
		"logic":  logic,
		"not":    not,
	})
	b.binding.AddBinding("where", parameters)
}

func (b *Builder) WhereBetween(column string, first, last interface{}) *Builder {
	if err := b.whereBetween(column, first, last, "and", false); err != nil {
		b.err = err
	}
	return b
}

func (b *Builder) WhereNotBetween(column string, first, last interface{}) *Builder {
	if err := b.whereBetween(column, first, last, "and", true); err != nil {
		b.err = err
	}
	return b
}

func (b *Builder) OrWhereBetween(column string, first, last interface{}) *Builder {
	if err := b.whereBetween(column, first, last, "or", false); err != nil {
		b.err = err
	}
	return b
}

func (b *Builder) OrWhereNotBetween(column string, first, last interface{}) *Builder {
	if err := b.whereBetween(column, first, last, "or", true); err != nil {
		b.err = err
	}
	return b
}

func (b *Builder) WhereNull(column string) *Builder {
	b.whereNull(column, "and", false)
	return b
}

func (b *Builder) WhereNotNull(column string) *Builder {
	b.whereNull(column, "and", true)
	return b
}

func (b *Builder) OrWhereNull(column string) *Builder {
	b.whereNull(column, "or", false)
	return b
}

func (b *Builder) OrWhereNotNull(column string) *Builder {
	b.whereNull(column, "or", true)
	return b
}

func (b *Builder) WhereSub(column, operator string, f func(b *Builder)) *Builder {
	b.whereSub(column, operator, "and", f)
	return b
}

func (b *Builder) OrWhereSub(column, operator string, f func(b *Builder)) *Builder {
	b.whereSub(column, operator, "or", f)
	return b
}

func (b *Builder) whereBetween(column string, first, last interface{}, logic string, not bool) (err error) {
	reft := reflect.TypeOf(first)
	switch reft.Kind() {
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		err = errors.New(fmt.Sprintf("whereQuery between first parameter need interface, not %s",
			reft.Kind().String()))
		return
	}
	reft = reflect.TypeOf(last)
	switch reft.Kind() {
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		err = errors.New(fmt.Sprintf("whereQuery between last parameter need interface, not %s",
			reft.Kind().String()))
		return
	}
	newval := []interface{}{first, last}
	b.wheres = append(b.wheres, map[string]interface{}{
		"type":   "Between",
		"column": column,
		"value":  newval,
		"logic":  logic,
		"not":    not,
	})
	b.binding.AddBinding("where", newval)
	return
}

func (b *Builder) whereNull(column, logic string, not bool) {
	b.wheres = append(b.wheres, map[string]interface{}{
		"type":   "Null",
		"column": column,
		"logic":  logic,
		"not":    not,
	})
}

func (b *Builder) whereSub(column, operator, logic string, f func(*Builder)) {
	builder := NewBuilder(b.driver, b.syntax, b.binding)
	f(builder)
	b.wheres = append(b.wheres, map[string]interface{}{
		"type":     "Sub",
		"column":   column,
		"operator": operator,
		"logic":    logic,
		"builder":  builder,
	})
}
