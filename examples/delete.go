package main

import (
	"github.com/Soul-Mate/sprydb"
	"log"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	manager := sprydb.NewManager()
	manager.AddConnection( "default",map[string]string{
		"username": "root",
		"password": "root",
		"host":     "127.0.0.1",
		"port":     "33060",
		"dbname":   "test",
		"driver":   "mysql",
	})
	if conn, err := manager.Connection("default"); err != nil {
		log.Fatal(err)
	} else {
		sql, err := conn.Table("users").
			Where("id", "=", 1).
			OrWhere("name", "=", "b").
			ToDeleteSql()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(sql)
	}
}
