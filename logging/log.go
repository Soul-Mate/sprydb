package logging

import (
	"bytes"
	"fmt"
)

type Logging struct {
	records []map[string]interface{}
}

func NewLogging() *Logging {
	return &Logging{}
}

func (l *Logging) Append(query string, bindings ...interface{}) {
	l.records = append(l.records, map[string]interface{}{
		"query":    query,
		"bindings": bindings,
	})
}

func (l *Logging) GetQueryLog() []map[string]interface{} {
	return l.records
}

func (l *Logging) GetRawQueryLog() []string {
	var (
		buf     bytes.Buffer
		query   string
		binding []interface{}
	)
	recordsLen := len(l.records)
	queries := make([]string, recordsLen)
	for i := 0; i < recordsLen; i++ {
		query = l.records[i]["query"].(string)
		binding = l.records[i]["bindings"].([]interface{})
		queryLen := len(query)
		bindingLen := len(binding)
		if bindingLen <= 0 {
			queries[i] = query
		} else {
			placeholderLen := 0
			for j := 0; j < queryLen; j++ {
				if query[j] == '?' {
					if placeholderLen < bindingLen {
						switch binding[placeholderLen].(type) {
						case string:
							buf.WriteString("\"%v\"")
						default:
							buf.WriteString("%v")
						}
					}
					placeholderLen++
				} else {
					buf.WriteByte(query[j])
				}
			}
			queries[i] = fmt.Sprintf(buf.String(), binding...)
			buf.Reset()
		}
	}
	return queries
}

func (l *Logging) GetQueryLogByIndex(index int) map[string]interface{} {
	if index > len(l.records) || index <= 0{
		return nil
	}
	return l.records[index-1]
}
