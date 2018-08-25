package syntax

import "github.com/Soul-Mate/sprydb/define"

type MysqlSyntax struct {
	whereOperators []string
	SyntaxAbstract
}

func NewMysqlSyntax() Syntax{
	return &MysqlSyntax{
		whereOperators:[]string{
			"=", "<", ">", "<=",
			">=", "<>", "!=", "<=>",
			"like", "like binary", "not like", "ilike",
			"&", "|", "^", "<<", ">>",
			"rlike", "regexp", "not regexp",
			"~", "~*", "!~", "!~*", "similar to",
			"not similar to", "not ilike", "~~*", "!~~*",
		},
	}
}

func (s *MysqlSyntax) PrepareWhereOperator(op string) (operator string, err error)  {
	if op == "" {
		return "=", nil
	}
	for _, v := range s.whereOperators {
		if v == op {
			return op, nil
		}
	}
	return "", define.InvalidOperatorError
	return
}