package tests

import (
	"testing"
	"github.com/Soul-Mate/sprydb"
	_"github.com/go-sql-driver/mysql"
)

func TestManager(t *testing.T)  {
	manager := sprydb.NewManager()
	configs := map[string]map[string]string{
		"master":{
			"username": "root",
			"password": "root",
			"host":     "127.0.0.1",
			"port":     "3306",
			"dbname":   "test",
			"driver":   "mysql",
			"connection_max":"1000",
			"connection_idle_max":"100",
			"connection_time":"60s",
		},
		"slave-1":{
			"username": "root",
			"password": "root",
			"host":     "127.0.0.1",
			"port":     "3306",
			"dbname":   "test",
			"driver":   "mysql",
			"connection_time":"60s",
		},
		"slave-2":{
			"username": "root",
			"password": "root",
			"host":     "127.0.0.1",
			"port":     "3306",
			"dbname":   "test",
			"driver":   "mysql",
			"connection_max":"1000",
			"connection_idle_max":"100",
			"connection_time":"60s",
		},
	}
	manager.AddMultiConnection(configs)
	for k := range configs {
		if conn, err := manager.Connection(k); err != nil {
			t.Error(err)
		} else {
			conn.Close()
		}

		// delete
		manager.DeleteConnection(k)

		// delete after get
		if conn, err := manager.Connection(k); err == nil {
			conn.Close()
			t.Error("DeleteConnection error")
		}
	}
}