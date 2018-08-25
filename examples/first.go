package main

import (
	"github.com/Soul-Mate/sprydb"
	"log"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Users struct {
	Id       int
	Name     string
	Password string
	Email    string
	Mobile   string
	Avatar   string
	Ip       string
	Token    string
}

func main() {
	manager := tinysql.NewManager()
	manager.AddConnection( "default",map[string]string{
		"username": "root",
		"password": "root",
		"host":     "127.0.0.1",
		"port":     "3306",
		"dbname":   "test",
		"driver":   "mysql",
	})
	if conn, err := manager.Connection("default"); err != nil {
		log.Fatal(err)
	} else {
		conn.EnableQueryLog()
		user := Users{}
		err := conn.First(&user)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(user)
		}
		fmt.Println(conn.GetQueryLogByIndex(1))
		user = Users{}
		result, err := conn.Table("users").Select("id", "name").FirstReturnMap()
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(result)
		}
	}
}
