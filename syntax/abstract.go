package syntax

import (
	"regexp"
	"strings"
	"reflect"
	"bytes"
)

type SyntaxAbstract struct {
}

func (s *SyntaxAbstract) ParseTable(str string) (table, alias string) {
	re, err := regexp.Compile(`(.*)\s+as\s+(.*)`)
	if err != nil {
		return str, ""
	}
	out := re.FindStringSubmatch(str)
	if len(out) < 3 {
		return str, ""
	}
	return out[1], out[2]
}

func (s *SyntaxAbstract) ParseColumn(str string) ()  {

}

func (s *SyntaxAbstract) WrapColumn(column string) (wrap string) {
	re, err := regexp.Compile(`(.*)\s+as\s+(.*)`)
	if err != nil {
		return column
	}
	ss := re.FindStringSubmatch(column)
	if n := len(ss); n >= 3 {
		return s.WrapColumn(ss[1]) + " as " + "`" + ss[2] + "`"
	}
	ss = strings.Split(column, ".")
	if n := len(ss); n >= 2 {
		return s.WrapColumn(ss[0]) + "." + "`" + ss[1] + "`"
	}
	return "`" + column + "`"
}

func (s *SyntaxAbstract) WrapTable(table string) (wrap string) {
	return "`" + table + "`"
}

func (s *SyntaxAbstract) WrapPrefixTable(prefix, table string) (wrap string) {
	if prefix == "" {
		return s.WrapTable(table)
	}
	return s.WrapTable(prefix + table)
}

func (s *SyntaxAbstract) WrapAliasTable(table, alias string) (wrap string) {
	if table == "" {
		return ""
	}
	if alias == "" {
		return s.WrapTable(table)
	}
	return s.WrapTable(table) + " as " + s.WrapTable(alias)
}

func (s *SyntaxAbstract) ColumnToString(column []string) (columnStr string) {
	if len(column) <= 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for _, v := range column {
		buf.WriteString(s.WrapColumn(v))
		buf.WriteString(",")
	}
	return buf.String()[:buf.Len()-1]
	return
}

func (s *SyntaxAbstract) ColumnToInsertString(column []string) (columnStr string) {
	if len(column) <= 0 {
		return
	}
	buf := bytes.Buffer{}
	for _, v := range column {
		buf.WriteString("`")
		buf.WriteString(v)
		buf.WriteString("`,")
	}
	columnStr = buf.String()[:buf.Len()-1]
	return
}

func (s *SyntaxAbstract) ColumnToUpdateString(column []string) (columnStr string) {
	buf := bytes.Buffer{}
	if len(column) <= 0 {
		return
	}
	for _, v := range column {
		buf.WriteString(s.WrapColumn(v))
		buf.WriteString(" = ?,")
	}
	columnStr = buf.String()[:buf.Len()-1]
	return
}

func (s *SyntaxAbstract) ParameterByLenToString(length int) (ParameterStr string) {
	switch length {
	case 0:
		return ""
	case 1:
		return "?"
	case 2:
		return "?,?"
	case 3:
		return "?,?,?"
	case 4:
		return "?,?,?,?"
	case 5:
		return "?,?,?,?,?"
	case 6:
		return "?,?,?,?,?,?"
	case 7:
		return "?,?,?,?,?,?,?"
	case 8:
		return "?,?,?,?,?,?,?,?"
	case 9:
		return "?,?,?,?,?,?,?,?,?"
	case 10:
		return "?,?,?,?,?,?,?,?,?,?"
	default:
		return strings.Repeat("?,", length)[:length*2-1]
	}
}

func (s *SyntaxAbstract) ParameterByInterfaceToString(any interface{}) (ParameterStr string) {
	switch any.(type) {
	case int, int8, int16, int32, int64,
	uint, uint8, uint16, uint32, uint64,
	float32, float64, string, bool:
		ParameterStr = "?"
	case []int, []int8, []int32, []int64,
	[]uint, []uint8, []uint16, []uint32, []uint64,
	[]float32, []float64, []string, []bool, []interface{}:
		n := reflect.ValueOf(any).Len()
		switch n {
		case 0:
			ParameterStr = ""
		case 1:
			ParameterStr = "?"
		case 2:
			ParameterStr = "?,?"
		case 3:
			ParameterStr = "?,?,?"
		case 4:
			ParameterStr = "?,?,?,?"
		case 5:
			ParameterStr = "?,?,?,?,?"
		case 6:
			ParameterStr = "?,?,?,?,?,?"
		case 7:
			ParameterStr = "?,?,?,?,?,?,?"
		case 8:
			ParameterStr = "?,?,?,?,?,?,?,?"
		case 9:
			ParameterStr = "?,?,?,?,?,?,?,?,?"
		case 10:
			ParameterStr = "?,?,?,?,?,?,?,?,?,?"
		default:
			ParameterStr = strings.TrimRight(
				strings.Repeat("?, ", reflect.ValueOf(any).Len()), ", ")
		}
	default:
		ParameterStr = ""
	}
	return
}

func (s *SyntaxAbstract) PrepareWhereOperator(op string) (operator string, err error) {
	return
}
