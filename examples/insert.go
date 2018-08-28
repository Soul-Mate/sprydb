package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
)

type UserProfileImpl struct {
	profile string
}

func (p *UserProfileImpl) ReadFromDB(data []byte) {
	p.profile = string(data)

}

func (p *UserProfileImpl) WriteToDB() []byte {
	return []byte(p.profile)
}

type Users struct {
	Id        int
	Name      string           `spry:"col:name"`
	CreatedAt time.Time        `spry:"col:created_at;use_alias:false;"`
	Profile   *UserProfileImpl `spry:"col:profile"`
}

func main() {
	manager := sprydb.NewManager()
	manager.AddConnection("default", map[string]string{
		"username": "root",
		"password": "root",
		"host":     "127.0.0.1",
		"port":     "33060",
		"dbname":   "test",
		"driver":   "mysql",
	})
	insertExample(manager)
	multiInsertExample(manager)
	transactionExample(manager)
}

func insertExample(manager *sprydb.Manager)  {
	conn, err := manager.Connection("default")
	if err != nil {
		log.Fatal(err)
	}
	user := Users{}
	user.Name = "user1"
	user.CreatedAt = time.Now()
	user.Profile = &UserProfileImpl{
		"profile....",
	}
	id, record, err := conn.Insert(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("inser %d record, id: %d\n", record, id)
	conn.Close()
}

func multiInsertExample(manager *sprydb.Manager)  {
	conn, err := manager.Connection("default")
	if err != nil {
		log.Fatal(err)
	}
	var users = []Users{
		{
			Name:"user2",
			CreatedAt:time.Now().Add(time.Minute),
			Profile: &UserProfileImpl{
				"user2 profile...",
			},
		},
		{
			Name:"user3",
			CreatedAt:time.Now().Add(time.Minute * 2),
			Profile: &UserProfileImpl{
				"user3 profile...",
			},
		},
		{
			Name:"user4",
			CreatedAt:time.Now().Add(time.Minute * 3),
			Profile: &UserProfileImpl{
				"user4 profile...",
			},
		},
	}
	id, record, err := conn.Insert(&users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("inser %d record, id: %d\n", record, id)
	conn.Close()
}

func transactionExample(manager *sprydb.Manager)  {
	conn, err := manager.Connection("default")
	if err != nil {
		log.Fatal(err)
	}
	session, err := conn.BeginTransaction()
	if err != nil {
		log.Fatal(err)
	}
	user := Users{}
	user.Name = "user1"
	user.CreatedAt = time.Now()
	user.Profile = &UserProfileImpl{
		"profile....",
	}
	id, record, err := session.Insert(&user)
	if err != nil {
		session.Rollback()
		log.Fatal(err)
	}
	fmt.Printf("inser %d record, id: %d\n", record, id)

	if err = session.Commit(); err != nil {
		log.Fatal(err)
	}

	if err = session.Close(); err != nil {
		log.Fatal(err)
	}

	if err = conn.Close(); err != nil {
		log.Fatal(err)
	}
}