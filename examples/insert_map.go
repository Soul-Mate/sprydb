package main

import (
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
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
	id, record, err := conn.Table("users").Insert(map[string]interface{}{
		"name":       "bar",
		"show":       1,
		"created_at": time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("insert %d record, id: %d\n", record, id)

	id, record, err = conn.Table("users").Insert([]map[string]interface{}{
		{
			"name":       "rop",
			"show":       1,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"name":       "john",
			"show":       0,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"name":       "rap",
			"show":       1,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("insert %d record, id: %d\n", record, id)
}
