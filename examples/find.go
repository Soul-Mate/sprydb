package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/Soul-Mate/sprydb"
	"fmt"
	"os"
	"time"
)

type UserProfileImpl struct {
	profile string
}

func (p *UserProfileImpl) ReadFromDB(data []byte) {
	p.profile = string(data)
}

func (p *UserProfileImpl) WriteToDB() []byte {
	return nil
}

type Users struct {
	Id    int
	Name  string `spry:"col:name"`
	CreatedAt time.Time       `spry:"col:created_at;use_alias:false;"`
	//Profile   *UserProfileImpl `spry:"col:profile"`
	Profile   []byte `spry:"col:profile"`
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
		fmt.Fprintf(os.Stderr, "connection error: %v\n", err)
		os.Exit(1)
	}

	users := &Users{}
	if err = conn.Find(1, users); err != nil {
		fmt.Fprintf(os.Stderr, "find query error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("query user: %v\n", users.Profile)

	if err = conn.Table("users as a").Find(2, users, "a.name", "a.profile", "created_at"); err != nil {
		fmt.Fprintf(os.Stderr, "connection error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("query user: %v\n", users)
}
