package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)

type User struct {
	Id   int    `spry:"column:id"`
	Name string `spry:"column:name"`
	Show string `spry:"show"`
}

func (u *User) Table() string {
	return "users"
}

type PostContent struct {
	data string
	len  int
}

type Posts struct {
	Id          int          `spry:"column:id"`
	UserId      int          `spry:"column:user_id"`
	PostContent *PostContent `spry:"column:content"`
}

func (p *Posts) Table() string {
	return "posts"
}

func (p *PostContent) Read(data []byte) {
	p.len = len(data)
	p.data = string(data)
}

func (p *PostContent) Write() []byte {
	return nil
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
	session, err := conn.BeginTransaction(0)
	if err != nil {
		log.Fatal(err)
	}
	conn.EnableQueryLog()
	user := User{}
	_, err = session.Find(1, &user, "id", "name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
	post := Posts{}
	_, err = session.Find(1, &post, "id", "user_id", "content")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(post)
}
