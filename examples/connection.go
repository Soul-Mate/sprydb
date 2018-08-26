package main

import (
	"github.com/Soul-Mate/sprydb"
	_ "github.com/go-sql-driver/mysql"
		"fmt"
	"os"
)

func main()  {
	var (
		err error
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
		fmt.Fprintf(os.Stderr, "get connection error: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "cannot db connection: %v\n", err)
			os.Exit(1)
		}
	}()
	fmt.Println(conn)
}