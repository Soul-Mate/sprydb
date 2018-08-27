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
	Name      string          `spry:"col:name"`
	CreatedAt time.Time       `spry:"col:created_at;use_alias:false;"`
	Profile   *UserProfileImpl `spry:"col:profile"`
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
		"port":     "3306",
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
	//id, record, err = conn.Insert([]Users{
	//	{
	//		Name: "Object1",
	//	},
	//	{
	//		Name: "Object2",
	//	},
	//	{
	//		Name: "Object3",
	//	},
	//})
	//fmt.Printf("inser %d record, id: %d\n", id, record)
}
