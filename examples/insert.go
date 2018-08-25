package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/Soul-Mate/sprydb/mapper"
	"fmt"
	"time"
)

type Users struct {
	Id        int
	Name      string
	Show      int
	Flag   int
	CreatedAt mapper.Time
}

func main() {
	var (
		err  error
		conn *tinysql.Connection
	)
	manager := tinysql.NewManager()
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
	user := Users{
	}
	id, record, err := conn.Insert(user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("inser %d record, id: %d\n", id, record)
	id, record, err = conn.Insert([]Users{
		{
			Name:"Object1",
			Show:0,
		},
		{
			Name:"Object2",
			Show:1,
		},
		{
			Name:"Object3",
			Show:1,
			CreatedAt: *mapper.NewTime(time.Now(), "2006-01-02 15:04:05"),
		},
	})
	fmt.Printf("inser %d record, id: %d\n", id, record)
}
