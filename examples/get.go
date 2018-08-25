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
	Password string `sprydb:"column:password"`
	Email    string `tintsql:"column:email"`
	Mobile   string `sprydb:"column:mobile"`
	Avatar   string `sprydb:"column:avatar"`
	Ip       string `sprydb:"column:ip"`
	Token    string `sprydb:"column:token"`
	UserLevels struct {
		Id      int
		LevelId int
		UserId  int
	} `sprydb:"-"`
}

func main() {
	manager := tinysql.NewManager()
	manager.AddConnection("default", map[string]string{
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
		var users []Users
		err := conn.
			Table("users as users").Get(&users)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(users)
		}
		if results, err := conn.
			Select("id", "name").GetReturnMap(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(results)
		}
	}
}
