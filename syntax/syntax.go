package syntax

type Syntax interface {
	ParseTable(string) (table, alias string)
	WrapColumn(column string) (wrap string)
	WrapTable(table string) (wrap string)
	WrapPrefixTable(prefix, table string) (wrap string)
	WrapAliasTable(table, alias string) (wrap string)
	ColumnToString([]string) string
	ColumnToInsertString([]string) string
	ColumnToUpdateString([]string) string
	ParameterByLenToString(int) string
	ParameterByInterfaceToString(interface{}) string
	PrepareWhereOperator(string) (operator string, err error)
}

func NewSyntax(driver string) Syntax {
	switch driver {
	case "mysql":
		return NewMysqlSyntax()
	default:
		return nil
	}
}
