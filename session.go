package sprydb

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Soul-Mate/sprydb/binding"
	"github.com/Soul-Mate/sprydb/define"
	"github.com/Soul-Mate/sprydb/mapper"
	"github.com/Soul-Mate/sprydb/query"
	"github.com/Soul-Mate/sprydb/syntax"
	"reflect"
	"time"
	"hash/crc32"
)

var (
	TransactionAlreadyUseErr = errors.New("The transaction already use, please commit or rollabck.")
)

type Session struct {
	err          error
	ctx          context.Context
	syntax       syntax.Syntax
	grammar      query.GrammarInterface
	binding      *binding.Binding
	stmtCache    map[uint32]*sql.Stmt
	connection   *Connection
	transaction  *Transaction
	queryBuilder *query.Builder
}

func NewSession(connection *Connection) *Session {
	session := new(Session)
	session.ctx = context.Background()
	session.syntax = syntax.NewSyntax(connection.driver)
	session.grammar = query.NewGrammarFactory(connection.driver, session.syntax, session.binding, connection.style)
	session.binding = binding.NewBinding()
	session.stmtCache = make(map[uint32]*sql.Stmt)
	session.connection = connection
	session.queryBuilder = query.NewBuilder(connection.driver, session.syntax, session.binding)
	return session
}

func (s *Session) Close() error {
	var err error
	// clean stmt cache
	for _, cache := range s.stmtCache {
		err = cache.Close()
	}
	return err
}

func (s *Session) BeginTransaction(timeout time.Duration) (err error) {
	if s.transaction != nil {
		return TransactionAlreadyUseErr
	}

	s.transaction = newTransaction(s.connection.DB, timeout)
	return s.transaction.begin()
}

func (s *Session) Commit() (err error) {
	return s.transaction.commit()
}

func (s *Session) Rollback() (err error) {
	if s.transaction != nil {
		err = s.transaction.rollback()
		s.transaction = nil
		return
	}
	return
}

func (s *Session) Table(tableName string) *Session {
	s.queryBuilder.Table(tableName)
	return s
}

func (s *Session) Distinct() *Session {
	s.queryBuilder.Distinct()
	return s
}

func (s *Session) Select(column ...string) *Session {
	s.queryBuilder.Select(column...)
	return s
}

func (s *Session) Join(table, first, operator, second string) *Session {
	s.queryBuilder.Join(table, first, operator, second)
	return s
}

func (s *Session) JoinClosure(table string, closure func(*query.BuilderJoin)) *Session {
	s.queryBuilder.JoinClosure(table, closure)
	return s
}

func (s *Session) LeftJoin(table, first, operator, second string) *Session {
	s.queryBuilder.LeftJoin(table, first, operator, second)
	return s
}

func (s *Session) LeftJoinClosure(table string, closure func(*query.BuilderJoin)) *Session {
	s.queryBuilder.LeftJoinClosure(table, closure)
	return s
}

func (s *Session) RightJoin(table, first, operator, second string) *Session {
	s.queryBuilder.RightJoin(table, first, operator, second)
	return s
}

func (s *Session) RightJoinClosure(table string, closure func(*query.BuilderJoin)) *Session {
	s.queryBuilder.RightJoinClosure(table, closure)
	return s
}

func (s *Session) Where(column, operator string, parameters interface{}) *Session {
	s.queryBuilder.Where(column, operator, parameters)
	return s
}

func (s *Session) OrWhere(column, operator string, parameters interface{}) *Session {
	s.queryBuilder.OrWhere(column, operator, parameters)
	return s
}

func (s *Session) WhereIn(column string, parameters interface{}) *Session {
	s.queryBuilder.WhereIn(column, parameters)
	return s
}

func (s *Session) WhereNotIn(column string, parameters interface{}) *Session {
	//s.builder.WhereIn(column, parameters)
	s.queryBuilder.WhereNotIn(column, parameters)
	return s
}

func (s *Session) OrWhereIn(column string, parameters interface{}) *Session {
	s.queryBuilder.OrWhereIn(column, parameters)
	return s
}

func (s *Session) OrWhereNotIn(column string, parameters interface{}) *Session {
	s.queryBuilder.OrWhereNotIn(column, parameters)
	return s
}

func (s *Session) WhereBetween(column string, first, last interface{}) *Session {
	s.queryBuilder.WhereBetween(column, first, last)
	return s
}

func (s *Session) WhereNotBetween(column string, first, last interface{}) *Session {
	s.queryBuilder.WhereNotBetween(column, first, last)
	return s
}

func (s *Session) OrWhereBetween(column string, first, last interface{}) *Session {
	s.queryBuilder.OrWhereBetween(column, first, last)
	return s
}

func (s *Session) OrWhereNotBetween(column string, first, last interface{}) *Session {
	s.queryBuilder.OrWhereNotBetween(column, first, last)
	return s
}

func (s *Session) WhereNull(column string) *Session {
	s.queryBuilder.WhereNull(column)
	return s
}

func (s *Session) WhereNotNull(column string) *Session {
	s.queryBuilder.WhereNotNull(column)
	return s
}

func (s *Session) OrWhereNull(column string) *Session {
	s.queryBuilder.OrWhereNull(column)
	return s
}

func (s *Session) OrWhereNotNull(column string) *Session {
	s.queryBuilder.OrWhereNotNull(column)
	return s
}

func (s *Session) WhereSub(column, operator string, f func(b *query.Builder)) *Session {
	s.queryBuilder.WhereSub(column, operator, f)
	return s
}

func (s *Session) OrWhereSub(column, operator string, f func(b *query.Builder)) *Session {
	s.queryBuilder.OrWhereSub(column, operator, f)
	return s
}

func (s *Session) OrderBy(column string, direction string) *Session {
	s.queryBuilder.OrderBy(direction, column)
	return s
}

func (s *Session) OrderByMulti(columns []string, direction string) *Session {
	s.queryBuilder.OrderByMulti(columns, direction)
	return s
}

func (s *Session) Skip(n int) *Session {
	s.queryBuilder.Skip(n)
	return s
}

func (s *Session) Take(n int) *Session {
	s.queryBuilder.Take(n)
	return s
}

func (s *Session) Find(id int, object interface{}, column ...string) error {
	var (
		objMapper *mapper.Mapper
		err       error
		stmt      *sql.Stmt
		rows      *sql.Rows
		sqlStr    string
		address   []interface{}
		columns   []string
		table     string
		alias     string
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return err
	}

	if object == nil {
		return define.ObjectNoneError
	}

	t := reflect.TypeOf(object)
	// 映射的对象必须是指针
	if t.Kind() != reflect.Ptr {
		return define.UnsupportedTypeError
	}

	// 指针解引用后必须是一个struct类型
	if t.Elem().Kind() != reflect.Struct {
		return define.UnsupportedTypeError
	}

	if objMapper, err = mapper.NewMapper(object, s.syntax, s.connection.style); err != nil {
		return err
	}

	// 获取查询的table
	buildTable := s.queryBuilder.GetTable()
	if buildTable != "" {
		table, alias = buildTable, s.queryBuilder.GetAlias()
		objMapper.SetTable(table)
		objMapper.SetAlias(alias)
	}

	// 解析映射对象
	if err = objMapper.Parse(); err != nil {
		return err
	}

	// 如果table为空,使用映射的table
	if buildTable == "" {
		table, alias = objMapper.GetTable(), objMapper.GetAlias()
	}

	// 用户选择查询的列,在使用时传入
	buildColumn := s.queryBuilder.GetColumn()
	if len(buildColumn) <= 0 {
		if len(column) <= 0 {
			columns, address = objMapper.GetColumnAndAddress()
		} else {
			address = objMapper.GetAddressByColumn(column)
			columns = column
		}
	} else {
		address = objMapper.GetAddressByColumn(buildColumn)
		columns = buildColumn
	}

	// builder find sql
	sqlStr = s.grammar.CompileFind(s.queryBuilder.GetDistinct(), columns, table, alias, objMapper.GetPK())
	println(sqlStr)
	// 追加查询sql日志
	if s.connection.logging != nil {
		defer s.connection.logging.Append(sqlStr, id)
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return err
	}

	if rows, err = s.query(stmt, id); err != nil {
		return err
	}

	if !rows.Next() {
		return nil
	}

	if err = rows.Scan(address...); err != nil {
		return err
	}

	// 赋值
	objMapper.AssignAddressValue()

	return nil
}

func (s *Session) FindReturnMap(id int, pk string, column ...string) (map[string]interface{}, error) {
	var (
		rows     *sql.Rows
		stmt     *sql.Stmt
		err      error
		distinct bool
		table    string
		alias    string
		columns  []string
		result   map[string]interface{}
	)

	defer s.resetBuilder()

	table = s.queryBuilder.GetTable()
	// not use table
	if table = s.queryBuilder.GetTable(); table == "" {
		return nil, define.TableNoneError
	}

	alias = s.queryBuilder.GetAlias()
	distinct = s.queryBuilder.GetDistinct()
	// pk empty
	if pk == "" {
		return nil, errors.New("primary key can not be empty!")
	}

	sqlStr := s.grammar.CompileFind(
		distinct,
		column,
		table,
		alias,
		pk)

	if stmt, err = s.prepare(sqlStr); err != nil {
		return nil, err
	}

	if rows, err = s.query(stmt, id); err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	if columns, err = rows.Columns(); err != nil {
		return nil, err
	}

	columnLen := len(columns)
	values := make([]interface{}, columnLen)
	address := make([]interface{}, columnLen)
	for i := 0; i < columnLen; i++ {
		address[i] = &values[i]
	}

	if err = rows.Scan(address...); err != nil {
		return nil, err
	}
	result = make(map[string]interface{})

	for i := 0; i < columnLen; i++ {
		if _, ok := values[i].([]byte); ok {
			result[columns[i]] = string(values[i].([]byte))
		} else {
			result[columns[i]] = values[i]
		}
	}
	return result, nil
}

func (s *Session) First(object interface{}, column ...string) error {
	var (
		err       error
		stmt      *sql.Stmt
		rows      *sql.Rows
		sqlStr    string
		address   []interface{}
		objMapper *mapper.Mapper
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return err
	}

	if object == nil {
		return define.ObjectNoneError
	}

	t := reflect.TypeOf(object)
	if t.Kind() != reflect.Ptr {
		return define.UnsupportedTypeError
	}

	if t.Elem().Kind() != reflect.Struct {
		return define.UnsupportedTypeError
	}

	if objMapper, err = mapper.NewMapper(object, s.syntax, s.connection.style); err != nil {
		return err
	}

	buildTable := s.queryBuilder.GetTable()
	buildAlias := s.queryBuilder.GetAlias()
	if buildTable != "" {
		objMapper.SetTable(buildTable)
		objMapper.SetAlias(buildAlias)
	}

	joinMap := s.queryBuilder.GetJoinMap()
	objMapper.SetJoinMap(&joinMap)
	if err = objMapper.Parse(); err != nil {
		return err
	}

	if buildTable == "" {
		s.queryBuilder.Table(objMapper.GetTable())
		s.queryBuilder.SetAlias(objMapper.GetAlias())
	}

	// connectionStmtCache column
	address = s.prepareGiveColumnMapper(objMapper, column...)

	s.queryBuilder.Skip(0).Take(1)

	if sqlStr, err = s.grammar.CompileSelect(s.queryBuilder); err != nil {
		return err
	}

	if s.connection.logging != nil {
		defer s.connection.logging.Append(sqlStr, s.binding.GetBindings()...)
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return err
	}

	if rows, err = s.query(stmt, s.binding.GetBindings()...); err != nil {
		return err
	}

	if !rows.Next() {
		return nil
	}

	if err = rows.Scan(address...); err != nil {
		return err
	}
	objMapper.AssignAddressValue()
	return nil
}

func (s *Session) FirstReturnMap() (map[string]interface{}, error) {
	var (
		rows    *sql.Rows
		stmt    *sql.Stmt
		err     error
		sqlStr  string
		columns []string
		result  map[string]interface{}
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return nil, err
	}

	if sqlStr, err = s.grammar.CompileSelect(s.queryBuilder); err != nil {
		return nil, err
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return nil, err
	}

	if rows, err = s.query(stmt, s.binding.GetBindings()...); err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	if columns, err = rows.Columns(); err != nil {
		return nil, err
	}
	columnLen := len(columns)
	values := make([]interface{}, columnLen)
	address := make([]interface{}, columnLen)
	for i := 0; i < columnLen; i++ {
		address[i] = &values[i]
	}

	if err = rows.Scan(address...); err != nil {
		return nil, err
	}

	result = make(map[string]interface{})
	for i := 0; i < columnLen; i++ {
		if _, ok := values[i].([]byte); ok {
			result[columns[i]] = string(values[i].([]byte))
		} else {
			result[columns[i]] = values[i]
		}
	}
	return result, nil
}

func (s *Session) Get(objects interface{}, column ...string) error {
	var (
		stmt         *sql.Stmt
		err          error
		reflectValue reflect.Value
		reflectType  reflect.Type
		obj          interface{}
		sqlStr       string
		rows         *sql.Rows
		address      []interface{}
		objMapper    *mapper.Mapper
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return err
	}

	reflectValue = reflect.ValueOf(objects)

	if reflectType = reflectValue.Type(); reflectType.Kind() != reflect.Ptr {
		return define.UnsupportedTypeError
	}

	if reflectType.Elem().Kind() != reflect.Slice {
		return errors.New("The method need slice type.")
	}

	if reflectType.Elem().Elem().Kind() != reflect.Struct {
		return errors.New("The elements in this slice should be struct.")
	}

	// create this type and get interface
	obj = reflect.New(reflectType.Elem().Elem()).Interface()

	if objMapper, err = mapper.NewMapper(obj, s.syntax, s.connection.style); err != nil {
		return err
	}

	buildTable := s.queryBuilder.GetTable()
	buildAlias := s.queryBuilder.GetAlias()
	if buildTable != "" {
		objMapper.SetTable(buildTable)
		objMapper.SetAlias(buildAlias)
	}

	joinMap := s.queryBuilder.GetJoinMap()
	objMapper.SetJoinMap(&joinMap)
	if err = objMapper.Parse(); err != nil {
		return err
	}

	if buildTable == "" {
		s.queryBuilder.Table(objMapper.GetTable())
		s.queryBuilder.SetAlias(objMapper.GetAlias())
	}

	address = s.prepareGiveColumnMapper(objMapper, column...)

	if sqlStr, err = s.grammar.CompileSelect(s.queryBuilder); err != nil {
		return err
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return err
	}

	if rows, err = s.query(stmt, s.binding.GetBindings()...); err != nil {
		return err
	}

	// TODO 是否需要清空传递的指针对象,确保不会对传递的slice进行追加
	elem := reflectValue.Elem()
	//newelem := reflect.New(reflectType.Elem()).Elem()
	for rows.Next() {
		if err = rows.Scan(address...); err != nil {
			return err
		}
		objMapper.AssignAddressValue()
		elem.Set(reflect.Append(elem, reflect.ValueOf(obj).Elem()))
		//newelem.Set(reflect.Append(newelem, reflect.ValueOf(obj).Elem()))
	}
	//elem.Set(newelem)
	return nil
}

func (s *Session) GetReturnMap() ([]map[string]interface{}, error) {
	var (
		err     error
		sqlStr  string
		stmt    *sql.Stmt
		rows    *sql.Rows
		columns []string
		results []map[string]interface{}
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return nil, err
	}

	if sqlStr, err = s.grammar.CompileSelect(s.queryBuilder); err != nil {
		return nil, err
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return nil, err
	}

	if rows, err = s.query(stmt, s.binding.GetBindings()...); err != nil {
		return nil, err
	}

	if columns, err = rows.Columns(); err != nil {
		return nil, err
	}
	columnLen := len(columns)
	values := make([]interface{}, columnLen)
	address := make([]interface{}, columnLen)
	for i := 0; i < columnLen; i++ {
		address[i] = &values[i]
	}
	for rows.Next() {
		if err = rows.Scan(address...); err != nil {
			continue
		}
		result := make(map[string]interface{})
		for i := 0; i < columnLen; i++ {
			if _, ok := values[i].([]byte); ok {
				result[columns[i]] = string(values[i].([]byte))
			} else {
				result[columns[i]] = values[i]
			}
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Session) prepareGiveColumnMapper(m *mapper.Mapper, column ...string) []interface{} {
	builderColumn := s.queryBuilder.GetColumn()
	if len(builderColumn) <= 0 {
		if len(column) <= 0 {
			col, addr := m.GetColumnAndAddress()
			s.queryBuilder.Select(col...)
			return addr
		} else {
			s.queryBuilder.Select(column...)
			addr := m.GetAddressByColumn(column)
			return addr
		}
	} else {
		addr := m.GetAddressByColumn(builderColumn)
		return addr
	}
}

func (s *Session) Insert(value interface{}) (lastInsertId, rowsAffected int64, err error) {
	var (
		stmt     *sql.Stmt
		sqlStr   string
		result   sql.Result
		bindings []interface{}
	)

	defer s.resetBuilder()

	if sqlStr, bindings, err = s.grammar.CompileInsert(value, s.queryBuilder); err != nil {
		return
	}

	if sqlStr == "" {
		return
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return
	}

	if result, err = s.exec(stmt, bindings...); err != nil {
		return
	}

	if lastInsertId, err = result.LastInsertId(); err != nil {
		return
	}

	if rowsAffected, err = result.RowsAffected(); err != nil {
		return
	}
	return
}

func (s *Session) Update(value interface{}) (rowsAffected int64, err error) {
	var (
		stmt     *sql.Stmt
		sqlStr   string
		result   sql.Result
		bindings []interface{}
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return 0, err
	}

	if sqlStr, bindings, err = s.grammar.CompileUpdate(value, s.queryBuilder); err != nil {
		return
	}

	if sqlStr == "" {
		return
	}
	bindings = s.binding.PrepareUpdateBinding(bindings)

	if stmt, err = s.prepare(sqlStr); err != nil {
		return
	}

	if result, err = s.exec(stmt, bindings...); err != nil {
		return
	}

	return result.RowsAffected()
}

func (s *Session) Delete() (rowsAffected int64, err error) {
	var (
		stmt   *sql.Stmt
		result sql.Result
		sqlStr string
	)

	defer s.resetBuilder()

	if err = s.queryBuilder.GetErr(); err != nil {
		return 0, err
	}

	if sqlStr, err = s.grammar.CompileDelete(s.queryBuilder); err != nil {
		return
	}

	if stmt, err = s.prepare(sqlStr); err != nil {
		return
	}

	if result, err = s.exec(stmt, s.binding.PrepareDeleteBinding()...); err != nil {
		return
	}

	return result.RowsAffected()
}

func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	stmt, err := s.prepare(query)
	if err != nil {
		return nil, err
	}
	return s.exec(stmt, args...)
}

func (s *Session) exec(stmt *sql.Stmt, bindings ...interface{}) (sql.Result, error) {
	if s.transaction != nil {
		return s.transaction.exec(stmt, bindings...)
	} else {
		return stmt.ExecContext(s.ctx, bindings...)
	}
}

func (s *Session) prepare(query string) (*sql.Stmt, error) {
	if stmt, err := s.connection.connectionStmtCache(query); err == nil {
		return stmt, nil
	}
	// cal check sum
	c32 := crc32.ChecksumIEEE([]byte(query))
	// load stmt stmtCache
	if stmtCache, ok := s.stmtCache[c32]; ok {
		return stmtCache, nil
	}
	if newStmt, err := s.connection.DB.PrepareContext(s.ctx, query); err != nil {
		return nil, err
	} else {
		s.stmtCache[c32] = newStmt
		return newStmt, nil
	}
}

func (s *Session) query(stmt *sql.Stmt, bindings ...interface{}) (rows *sql.Rows, err error) {
	if s.transaction != nil {
		return s.transaction.query(stmt, bindings...)
	} else {
		return stmt.QueryContext(s.ctx, bindings...)
	}
}

func (s *Session) ToSql() (sql string, err error) {
	defer s.resetBuilder()

	return s.grammar.CompileSelect(s.queryBuilder)
}

// TODO 集成log
func (s *Session) ToInsertSql(bean interface{}) (sql string, err error) {
	defer s.resetBuilder()
	sql, _, err = s.grammar.CompileInsert(bean, s.queryBuilder)
	return
}

func (s *Session) ToUpdateSql(bean interface{}) (sql string, err error) {
	defer s.resetBuilder()
	sql, _, err = s.grammar.CompileUpdate(bean, s.queryBuilder)
	return
}

func (s *Session) ToDeleteSql() (sql string, err error) {
	defer s.resetBuilder()
	sql, err = s.grammar.CompileDelete(s.queryBuilder)
	return
}

func (s *Session) resetBuilder() {
	s.queryBuilder = query.NewBuilder(s.connection.driver, s.syntax, binding.NewBinding())
}
