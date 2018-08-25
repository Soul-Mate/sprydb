package tests

import (
	"testing"
	"github.com/Soul-Mate/sprydb"
	"sync"
	"github.com/Soul-Mate/sprydb/logging"
	"unsafe"
	"github.com/pkg/errors"
)

var manager = tinysql.NewManager()


func Test_EnableQueryLog(t *testing.T) {
	manager.AddConnection("default", map[string]string{
		"username":            "root",
		"password":            "root",
		"host":                "127.0.0.1",
		"port":                "3306",
		"dbname":              "test",
		"driver":              "mysql",
		"connection_max":      "1000",
		"connection_idle_max": "100",
		"connection_time":     "60s",
	})
	conn, err := manager.Connection("default")
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	wg := sync.WaitGroup{}
	var local []*logging.Logging
	for i := 0; i < 100; i ++ {
		wg.Add(1)
		go func() {
			local = append(local, conn.EnableQueryLog())
			wg.Done()
		}()
	}
	wg.Wait()
	for _, v := range local {
		for _, vv := range local {
			if unsafe.Pointer(vv) != unsafe.Pointer(v) {
				t.Error(errors.New("testEnableQueryLog error"))
			}
		}
	}
}

func Test_GetQueryLog(t *testing.T)  {
	manager.AddConnection("default", map[string]string{
		"username":            "root",
		"password":            "root",
		"host":                "127.0.0.1",
		"port":                "3306",
		"dbname":              "test",
		"driver":              "mysql",
		"connection_max":      "1000",
		"connection_idle_max": "100",
		"connection_time":     "60s",
	})
	conn, err := manager.Connection("default")
	if err != nil {
		t.Error(err)
	}
	conn.EnableQueryLog()
	sql1, err := conn.Table("users").Where("id", "=", 1).
		OrderBy("id", "desc").
		OrWhere("id", "=", 2).
		WhereSub("id", "in", func(b *tinysql.Builder) {
		b.Table("users").Select("id")
	}).ToSql()
	sql2, err := conn.Table("users").Where("id", "=", 1).
		OrderBy("id", "desc").
		OrWhere("id", "=", 2).ToSql()
	logs := conn.GetQueryLog()
	if len(logs) != 2{
		t.Error("GetQueryLog error")
	}
	if logs[0] == nil || logs[1] == nil{
		t.Error("GetQueryLog error")
	}
	if logs[0]["query"] != sql1 {
		t.Error("GetQueryLog error")
	}
	if logs[1]["query"] != sql2 {
		t.Error("GetQueryLog error")
	}
}

func Test_GetQueryLogByIndex(t *testing.T) {
	manager.AddConnection("default", map[string]string{
		"username":            "root",
		"password":            "root",
		"host":                "127.0.0.1",
		"port":                "3306",
		"dbname":              "test",
		"driver":              "mysql",
		"connection_max":      "1000",
		"connection_idle_max": "100",
		"connection_time":     "60s",
	})
	conn, err := manager.Connection("default")
	if err != nil {
		t.Error(err)
	}
	conn.EnableQueryLog()
	sql, err := conn.Table("users").Where("id", "=", 1).
		OrderBy("id", "desc").
		OrWhere("id", "=", 2).
		WhereSub("id", "in", func(b *tinysql.Builder) {
		b.Table("users").Select("id")
	}).ToSql()
	if err != nil {
		t.Error(err)
	}
	log := conn.GetQueryLogByIndex(1)
	if log == nil || len(log)  <= 0 {
		t.Error("GetQueryLogByIndex error")
	}
	if log["query"] != sql {
		t.Error("log record sql not equal")
	}
	if log["bindings"].([]interface{})[0] != 1 || log["bindings"].([]interface{})[1] != 2 {
		t.Error("log record binding not equal")
	}
}

