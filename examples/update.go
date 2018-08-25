package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/Soul-Mate/sprydb/mapper"
)

type Users struct {
	Id        int
	Name      string
	Show      int `spry:"update_zero:true"`
	Flag      int
	CreatedAt mapper.Time
}

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
	user := Users{
		Name: "sprydb",
		Show: 0,
	}
	_, err = conn.Where("id", "=", 1).Update(user)
	if err != nil {
		log.Fatal(err)
	}
}
