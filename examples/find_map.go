package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

func main() {
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
	conn.EnableQueryLog()
	result, err := conn.Table("users").FindReturnMap(1, "id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
	result, err = conn.Table("user_posts").
		FindReturnMap(1, "id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
