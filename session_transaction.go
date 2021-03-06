package sprydb

import (
	"database/sql"
	"context"
)

type Transaction struct {
	db      *sql.DB
	tx      *sql.Tx
	ctx     context.Context
	cancel  context.CancelFunc
}

func newTransaction(db *sql.DB) *Transaction {
	transaction := new(Transaction)
	transaction.db = db
	transaction.tx = nil
	transaction.ctx = context.Background()
	return transaction
}

func (t *Transaction) begin() (err error) {
	var tx  *sql.Tx
	// 设置了事务超时时间
	// 事务使用新的context来进行超时管理
	//if t.timeout != 0 {
	//	t.ctx, t.cancel = context.WithTimeout(t.ctx, t.timeout)
	//}
	if tx, err = t.db.BeginTx(t.ctx, nil); err != nil {
		return
	}
	t.tx = tx
	return
}

func (t *Transaction) queryRow(stmt *sql.Stmt, bindings ...interface{}) *sql.Row {
	txStmt := t.tx.Stmt(stmt)
	return txStmt.QueryRowContext(t.ctx, bindings)
}

func (t *Transaction) query(stmt *sql.Stmt, bindings ...interface{}) (*sql.Rows, error) {
	txStmt := t.tx.Stmt(stmt)
	return txStmt.QueryContext(t.ctx, bindings...)
}

func (t *Transaction) exec(stmt *sql.Stmt, bindings ...interface{}) (sql.Result, error) {
	txStmt := t.tx.Stmt(stmt)
	result, err := txStmt.ExecContext(t.ctx, bindings...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Transaction) commit() (err error) {
	return t.tx.Commit()
}

func (t *Transaction) rollback() (err error) {
	return t.tx.Rollback()
}
