package main

import (
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"fmt"
)

func main()  {
	var (
		err  error
		conn *sprydb.Connection
	)
	manager := sprydb.NewManager()
	manager.AddConnection("default", map[string]string{
		"username": "root",
		"password": "root",
		"host":     "127.0.0.1",
		"port":     "33060",
		"dbname":   "test",
		"driver":   "mysql",
	})
	if conn, err = manager.Connection("default"); err != nil {
		log.Fatal(err)
	}
	session, err := conn.BeginTransaction(0)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := recover(); err != nil {
			session.Rollback()
			log.Fatal(err)
		}
	}()
	row, err := session.Table("users").Where("id", "=", 1).Update(map[string]interface{}{
		"name":"like:spry-sql",
		"show":1,
		"flag":1,
	})
	if err != nil {
		session.Rollback()
		log.Fatal(err)
	}
	fmt.Println("row affected: ", row)
	row, err = session.Table("users").Where("id", ">", 1).Delete()
	if err != nil {
		session.Rollback()
		log.Fatal(err)
	}
	fmt.Println("row affected: ", row)
	panic("line 41")
	session.Commit()
}