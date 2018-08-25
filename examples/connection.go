package main

import (
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"fmt"
)

func main()  {
	var (
		err error
		conn *sprydb.Connection
	)
	manager := sprydb.NewManager()
	manager.AddConnection("default", map[string]string{
		"username": "root",
		"password": "root",
		"host":     "127.0.0.1",
		"port":     "3306",
		"dbname":   "test",
		"driver":   "mysql",
	})
	if conn, err = manager.Connection("default"); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	fmt.Println(conn)
}