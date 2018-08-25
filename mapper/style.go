package mapper

import (
	"unicode"
	"bytes"
)

// 映射器的风格接口
type MapperStyler interface {
	table(string) string
	column(string) string
}

type CamelMapperStyle struct{}

func (*CamelMapperStyle) table(str string) string {
	var buf bytes.Buffer
	strLen := len(str)
	buf.WriteRune(unicode.ToUpper(rune(str[0])))
	for i := 1; i < strLen; i++ {
		switch str[i] {
		case '_', ' ':
			if i+1 < strLen {
				buf.WriteRune(unicode.ToUpper(rune(str[i+1])))
			}
			i++
		default:
			buf.WriteByte(str[i])
		}
	}
	return buf.String()
}

func (*CamelMapperStyle) column(str string) string {
	var buf bytes.Buffer
	strLen := len(str)
	buf.WriteRune(unicode.ToUpper(rune(str[0])))
	for i := 1; i < strLen; i++ {
		switch str[i] {
		case '_', ' ':
			if i+1 < strLen {
				buf.WriteRune(unicode.ToUpper(rune(str[i+1])))
			}
			i++
		default:
			buf.WriteByte(str[i])
		}
	}
	return buf.String()
}

type UnderlineMapperStyle struct{}

func (*UnderlineMapperStyle) table(str string) string {
	var buf bytes.Buffer
	buf.WriteRune(unicode.ToLower(rune(str[0])))
	for i, n := 1, len(str); i < n; i++ {
		r := rune(str[i])
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
			buf.WriteString("_")
			buf.WriteRune(r)
		} else {
			buf.WriteByte(str[i])
		}
	}
	return buf.String()
}

func (*UnderlineMapperStyle) column(str string) string {
	var buf bytes.Buffer
	buf.WriteRune(unicode.ToLower(rune(str[0])))
	for i, n := 1, len(str); i < n; i++ {
		r := rune(str[i])
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
			buf.WriteString("_")
			buf.WriteRune(r)
		} else {
			buf.WriteByte(str[i])
		}
	}
	return buf.String()
}
