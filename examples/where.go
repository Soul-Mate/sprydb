package main

import (
	"log"
		_ "github.com/go-sql-driver/mysql"
	"fmt"
	"github.com/Soul-Mate/sprydb/mapper"
	"github.com/Soul-Mate/sprydb"
)

type User struct {
	Id        int    `spry:"column:id"`
	Name      string `spry:"column:name"`
	Show      string `spry:"show"`
	CreatedAt *mapper.Time
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
	PostContent PostContent `spry:"column:content"`
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
	conn.EnableQueryLog()
	user := User{}
	err = conn.Where("id", "=", "1").First(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
	post := Posts{}
	err = conn.Where("user_id", "=", user.Id).First(&post)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(post)
}
