package query

import (
	"github.com/Soul-Mate/sprydb/syntax"
	"github.com/Soul-Mate/sprydb/binding"
	"github.com/Soul-Mate/sprydb/mapper"
)

type MysqlGrammar struct {
	*Grammar
}

func NewMysqlGrammar(syntax syntax.Syntax, binding *binding.Binding, styler mapper.MapperStyler) *MysqlGrammar {
	return &MysqlGrammar{
		NewGrammar(syntax, binding, styler),
	}
}
