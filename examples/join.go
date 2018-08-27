package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	"github.com/Soul-Mate/sprydb/query"
)

type Users struct {
	Id        int             `spry:"col:id"`
	Name      string          `spry:"col:name"`
	CreatedAt time.Time       `spry:"col:created_at;;"`
	Profile   UserProfileImpl `spry:"col:profile"`
	//Posts     *UserPosts      `spry:"extend:user_posts"`
}

type UserProfileImpl struct {
	profile string
}

func (p *UserProfileImpl) ReadFromDB(data []byte) {
	p.profile = string(data)
}

func (p *UserProfileImpl) WriteToDB() []byte {
	return nil
}

type UserPosts struct {
	Id     int      `spry:"col:id"`
	UserId int      `spry:"col:user_id"`
	Post   PostImpl `spry:"col:post"`
}

type PostImpl struct {
	data string
	l    int
}

func (p *PostImpl) ReadFromDB(data []byte) {
	p.data = string(data)
	p.l = len(p.data)
}

func (p *PostImpl) WriteToDB() []byte {
	return []byte(p.data)
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
	users := &struct {
		Users     `spry:"extend"`
		UserPosts `spry:"extend"`
	}{}
	err = conn.Table("users as a").
		Join("user_posts as b", "a.id", "=", "b.user_id").
		First(users)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(users)

	var usersS []struct {
		Users     `spry:"extend"`
		UserPosts `spry:"extend"`
	}
	err = conn.Table("users as a").
		JoinClosure("user_posts as b", func(join *query.BuilderJoin) {
			join.On("a.id", "=", "b.user_id").Where("a.id", "=", 1)
		}).Get(&usersS)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(usersS)
}
