package query


func (b *Builder) OrderBy(column string, direction string) *Builder {
	b.orderBy(direction, column)
	return b
}

func (b *Builder) OrderByMulti(columns []string, direction string) *Builder {
	b.orderBy(direction, columns...)
	return b
}

func (b *Builder) orderBy(direction string, col ...string) {
	switch direction {
	case "asc", "desc":
	case "ASC":
		direction = "asc"
	case "DESC":
		direction = "desc"
	default:
		direction = "asc"
	}
	b.orders["direction"] = direction
	b.orders["column"] = append(b.orders["column"].([]string), col...)
}
