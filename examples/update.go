package main

import (
	"log"
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"fmt"
	"os"
	"strconv"
	"sync"
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
		"port":     "3306",
		"dbname":   "test",
		"driver":   "mysql",
	})
	updateExample(manager)
	transactionUpdateExample(manager)
	concurrentTransactionUpdateExample(manager)
}

func updateExample(manager *sprydb.Manager) {
	conn, err := manager.Connection("default")
	if err != nil {
		log.Fatal(err)
	}
	user := Users{
		Name:      "spry db",
		CreatedAt: time.Now(),
		Profile: &UserProfileImpl{
			"update user profile...",
		},
	}
	_, err = conn.Where("id", "=", 1).Update(user)
	if err != nil {
		log.Fatal(err)
	}
}

func transactionUpdateExample(manager *sprydb.Manager) {
	conn, err := manager.Connection("default")
	if err != nil {
		log.Fatal(err)
	}

	// 开启事务
	session, err := conn.BeginTransaction()
	if err != nil {
		log.Fatal(err)
	}

	user := Users{
		Name:      "spry-db",
		CreatedAt: time.Now().Add(time.Hour),
		Profile: &UserProfileImpl{
			"update user profile in transaction...",
		},
	}
	_, err = session.Where("id", "=", 1).Update(user)
	if err != nil {
		log.Fatal(err)
	}

	if err = session.Commit(); err != nil {
		log.Fatal(err)
	}

	session.Close()
	conn.Close()
}

func concurrentTransactionUpdateExample(manager *sprydb.Manager) {
	conn, err := manager.Connection("default")
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	// 开启10个线程
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			// 开启事务
			session, err := conn.BeginTransaction()
			fmt.Printf("trsanction-%d start.\n", n)
			if err != nil {
				fmt.Fprintf(os.Stderr, "trsanction error-%d\n", n)
			}
			user := Users{
				Name:      "spry-" + strconv.Itoa(n),
				CreatedAt: time.Now().Add(time.Hour),
				Profile: &UserProfileImpl{
					"update user profile in transaction...",
				},
			}
			_, err = session.Where("id", "=", 1).Update(user)
			if err != nil {
				log.Fatal(err)
			}

			time.AfterFunc(time.Second , func() {
				if err = session.Commit(); err != nil {
					log.Fatal(err)
				}
			})


			session.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()
}
