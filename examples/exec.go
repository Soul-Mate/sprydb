package main

import (
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"fmt"
	"io/ioutil"
	"sync"
	"strconv"
	"time"
)

func main() {
	var (
		err  error
		conn *sprydb.Connection
	)

	const SQL_PATH = "/Users/xyc/Code/go/src/github.com/Soul-Mate/sprydb/examples/example.sql"
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

	defer func() {
		if err = conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	f, err := os.Open(SQL_PATH)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open file: %v\n", err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "cannot close file: %v\n", err)
		}
	}()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read the file: %v\n", err)
		os.Exit(1)
	}

	if _, err = conn.Exec(string(b)); err != nil {
		fmt.Fprintf(os.Stderr, "cannot exec sql: %v\n", err)
		os.Exit(1)
	}

	insert(100, conn)
}

func insert(n int, conn *sprydb.Connection) {
	wg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		// 并发事务执行
		go func(v int) {
			defer wg.Done()

			session, err := conn.BeginTransaction(0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "transaction error: %v\n", err)
				return
			}

			defer func() {
				if err = session.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "session close error: %v\n", err)
				}
			}()

			insertSQL := "INSERT INTO users (name, created_at, profile) VALUES (?, ?, ?);"
			result, err := session.Exec(insertSQL, "user"+strconv.Itoa(v), time.Now(), "this is the user profile")
			if err != nil {
				fmt.Fprintf(os.Stderr, "inser user error: %v", err)
				if err = session.Rollback(); err != nil {
					fmt.Fprintf(os.Stderr, "session rollback error: %v", err)
				}
				return
			}

			id, _ := result.LastInsertId()
			fmt.Fprintf(os.Stdout, "insert user success: %d\n", id)
			if err = session.Commit(); err != nil {
				fmt.Fprintf(os.Stderr, "session rollback error: %v", err)
				return
			}
		}(i)
	}
	wg.Wait()
}
