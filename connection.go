package sprydb

import (
	"database/sql"
	"errors"
	"sync"
	"github.com/Soul-Mate/sprydb/logging"
	"time"
	"fmt"
	"strconv"
		"github.com/Soul-Mate/sprydb/query"
	"github.com/Soul-Mate/sprydb/mapper"
	"hash/crc32"
)

var (
	once                     = sync.Once{}
	missingUserNameErr       = errors.New("The Connection missing parameter: username.")
	missingPasswordErr       = errors.New("The Connection missing parameter: password.")
	missingHostErr           = errors.New("The Connection missing parameter: host.")
	missingPortErr           = errors.New("The Connection missing parameter: port.")
	missingDBNameErr         = errors.New("The Connection missing parameter: dbname.")
	setConnectionLifeTimeErr = errors.New("The connection_time parameter format error.")
)

type Connection struct {
	Err     error
	driver  string
	name    string
	DB      *sql.DB
	cache   *sync.Map
	logging *logging.Logging
	style   mapper.MapperStyler
}

func NewConnection(config map[string]string) (*Connection, error) {
	var (
		db             *sql.DB
		err            error
		dataSourceName string
	)
	if dataSourceName, err = buildConnectionSource(config); err != nil {
		return nil, err
	}

	if db, err = sql.Open(config["driver"], dataSourceName); err != nil {
		return nil, err
	}

	// parse and set connection max life time
	if v, ok := config["connection_time"]; ok {
		var connLiftTime time.Duration
		if timeDuration, err := time.ParseDuration(v); err != nil {
			if timeInt, err := strconv.Atoi(v); err != nil {
				db.Close()
				return nil, setConnectionLifeTimeErr
			} else {
				connLiftTime = time.Duration(timeInt)
			}
		} else {
			connLiftTime = timeDuration
		}
		db.SetConnMaxLifetime(connLiftTime)
	}

	// parse and set connection pool max num
	if v, ok := config["connection_max"]; ok {
		if max, err := strconv.Atoi(v); err != nil {
			db.Close()
			return nil, setConnectionLifeTimeErr
		} else {
			db.SetMaxOpenConns(max)
		}
	}

	// parse and set idle connection pool num
	if v, ok := config["connection_idle_max"]; ok {
		if max, err := strconv.Atoi(v); err != nil {
			db.Close()
			return nil, setConnectionLifeTimeErr
		} else {
			db.SetMaxOpenConns(max)
		}
	}

	conn := new(Connection)
	conn.DB = db
	conn.cache = new(sync.Map)
	conn.logging = logging.NewLogging()
	conn.driver = config["driver"]
	return conn, nil
}

func buildConnectionSource(config map[string]string) (source string, err error) {
	var (
		ok                                     bool
		username, password, host, port, dbname string
	)
	if username, ok = config["username"]; !ok {
		err = missingUserNameErr
	}

	if password, ok = config["password"]; !ok {
		err = missingPasswordErr
	}

	if host, ok = config["host"]; !ok {
		err = missingHostErr
	}

	if port, ok = config["port"]; !ok {
		err = missingPortErr
	}

	if dbname, ok = config["dbname"]; !ok {
		err = missingDBNameErr
	}
	if err != nil {
		return
	}
	source = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)
	return
}

func (c *Connection) SetMapperStyle(style mapper.MapperStyler) {
	c.style = style
}

// 关闭数据库连接
func (c *Connection) Close() error {
	var err error
	// 清理stmt cache
	c.cache.Range(func(key, value interface{}) bool {
		if err = value.(*sql.Stmt).Close(); err != nil {
			return false
		}
		c.cache.Delete(key)
		return true
	})
	err = c.DB.Close()
	return err
}

// 执行原始sql
func (c *Connection) Exec(query string, args ...interface{}) (sql.Result, error) {
	session := NewSession(c)
	return session.Exec(query, args...)
}

// 最上级的stmt cache
func (c *Connection) connectionStmtCache(query string) (*sql.Stmt, error) {
	c32 := crc32.ChecksumIEEE([]byte(query))
	if cache, ok := c.cache.Load(c32); ok {
		return cache.(*sql.Stmt), nil
	}

	newStmt, err := c.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	c.cache.Store(c32, newStmt)
	return newStmt, nil
}

// 开启事务
func (c *Connection) BeginTransaction(timeout time.Duration) (*Session, error) {
	session := NewSession(c)
	err := session.BeginTransaction(timeout)
	return session, err
}

func (c *Connection) Table(tableName string) *Session {
	session := NewSession(c)
	session.Table(tableName)
	return session
}

func (c *Connection) Find(id int, object interface{}, column ...string) (err error) {
	session := NewSession(c)
	return session.Find(id, object, column...)
}

func (c *Connection) Select(column ...string) *Session {
	session := NewSession(c)
	session.Select(column...)
	return session
}

func (c *Connection) Join(table, first, operator, second string) *Session {
	session := NewSession(c)
	session.Join(table, first, operator, second)
	return session
}

func (c *Connection) JoinClosure(table string, closure func(*query.BuilderJoin)) *Session {
	session := NewSession(c)
	session.JoinClosure(table, closure)
	return session
}

func (c *Connection) LeftJoin(table, first, operator, second string) *Session {
	session := NewSession(c)
	session.LeftJoin(table, first, operator, second)
	return session
}

func (c *Connection) LeftJoinClosure(table string, closure func(*query.BuilderJoin)) *Session {
	session := NewSession(c)
	session.LeftJoinClosure(table, closure)
	return session
}

func (c *Connection) RightJoin(table, first, operator, second string) *Session {
	session := NewSession(c)
	session.RightJoin(table, first, operator, second)
	return session
}

func (c *Connection) RightJoinClosure(table string, closure func(*query.BuilderJoin)) *Session {
	session := NewSession(c)
	session.RightJoinClosure(table, closure)
	return session
}

func (c *Connection) Where(column, operator string, parameters interface{}) *Session {
	session := NewSession(c)
	session.Where(column, operator, parameters)
	if session.err != nil {
		c.Err = session.err
	}
	return session
}

func (c *Connection) OrWhere(column, operator string, parameters interface{}) *Session {
	session := NewSession(c)
	session.OrWhere(column, operator, parameters)
	if session.err != nil {
		c.Err = session.err
	}
	return session
}

func (c *Connection) WhereIn(column string, parameters interface{}) *Session {
	session := NewSession(c)
	session.WhereIn(column, parameters)
	return session
}

func (c *Connection) WhereNotIn(column string, parameters interface{}) *Session {
	session := NewSession(c)
	session.WhereNotIn(column, parameters)
	return session
}

func (c *Connection) OrWhereIn(column string, parameters interface{}) *Session {
	session := NewSession(c)
	session.OrWhereIn(column, parameters)
	return session
}

func (c *Connection) OrWhereNotIn(column string, parameters interface{}) *Session {
	session := NewSession(c)
	session.OrWhereNotIn(column, parameters)
	return session
}

func (c *Connection) WhereBetween(column string, first interface{}, last interface{}) *Session {
	session := NewSession(c)
	session.WhereBetween(column, first, last)
	return session
}

func (c *Connection) WhereNotBetween(column string, first, last interface{}) *Session {
	session := NewSession(c)
	session.WhereNotBetween(column, first, last)
	return session
}

func (c *Connection) OrWhereBetween(column string, first, last interface{}) *Session {
	session := NewSession(c)
	session.OrWhereBetween(column, first, last)
	return session
}

func (c *Connection) OrWhereNotBetween(column string, first, last interface{}) *Session {
	session := NewSession(c)
	session.OrWhereNotBetween(column, first, last)
	return session
}

func (c *Connection) WhereNull(column string) *Session {
	session := NewSession(c)
	session.WhereNull(column)
	return session
}

func (c *Connection) WhereNotNull(column string) *Session {
	session := NewSession(c)
	session.WhereNotNull(column)
	return session
}

func (c *Connection) OrWhereNull(column string) *Session {
	session := NewSession(c)
	session.OrWhereNull(column)
	return session
}

func (c *Connection) OrWhereNotNull(column string) *Session {
	session := NewSession(c)
	session.OrWhereNotNull(column)
	return session
}

func (c *Connection) OrderBy(column string, direction string) *Session {
	session := NewSession(c)
	session.OrderBy(column, direction)
	return session
}

func (c *Connection) OrderByMulti(columns []string, direction string) *Session {
	session := NewSession(c)
	session.OrderByMulti(columns, direction)
	return session
}

func (c *Connection) Skip(n int) *Session {
	session := NewSession(c)
	session.Skip(n)
	return session
}

func (c *Connection) Take(n int) *Session {
	session := NewSession(c)
	session.Take(n)
	return session
}

func (c *Connection) ToSql() (sqlString string, err error) {
	session := NewSession(c)
	if sqlString, err = session.ToSql(); err != nil {
		return
	}
	return
}

func (c *Connection) First(object interface{}, column ...string) error {
	session := NewSession(c)
	if err := session.First(object, column...); err != nil {
		return err
	}
	return nil
}

func (c *Connection) Insert(value interface{}) (lastInsertId, rowsAffected int64, err error) {
	session := NewSession(c)
	return session.Insert(value)
}

func (c *Connection) Update(value interface{}) (rowsAffected int64, err error) {
	session := NewSession(c)
	return session.Update(value)
}

func (c *Connection) EnableQueryLog() *logging.Logging {
	once.Do(func() {
		if c.logging == nil {
			c.logging = logging.NewLogging()
		}
	})
	return c.logging
}

func (c *Connection) GetQueryLog() []map[string]interface{} {
	return c.logging.GetQueryLog()
}

func (c *Connection) GetRawQueryLog() []string {
	return c.logging.GetRawQueryLog()
}

func (c *Connection) GetQueryLogByIndex(index int) map[string]interface{} {
	return c.logging.GetQueryLogByIndex(index)
}
