package mysql

import (
	"testing"
	"time"
)

func TestTableCmd(t *testing.T) {
	type tabledata struct {
		Id   int       `db:"id"`
		Name string    `db:"name"`
		Age  int       `db:"age"`
		Time time.Time `db:"time"`
	}
	if cmd, err := createTableCmd("aa", tabledata{}, ifnotOn); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Log(cmd)
	}
	if cmd, err := dropTableCmd("aa"); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Log(cmd)
	}
	if cmd, err := readTableAllCmd(); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Log(cmd)
	}
	if cmd, err := readCreateTableCmd("aa"); err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Log(cmd)
	}

}

func TestCreateDbTable(t *testing.T) {
	type tabledata struct {
		Id   int       `db:"id"`
		Name string    `db:"name"`
		Age  int       `db:"age"`
		Time time.Time `db:"time"`
	}
	t.Log("--------------- create Table -----------------")
	connect := CreateConectName("root", "root", "tcp", "localhost", "3306")
	if cfg, err := Setup(connect, "db_write"); err != nil {
		t.Fatalf(err.Error())
	} else {

		defer stopdb(cfg, connect, "db_write")
		if err := cfg.CreateTable("user", tabledata{}); err != nil {
			t.Fatalf(err.Error())
		}
		t.Log("--------------- create Table pass -----------------")
		if cmd, err := cfg.ReadCreateTableCmd("user"); err != nil {
			t.Fatalf(err.Error())
		} else {
			t.Log(cmd)
		}
		if err := cfg.DropTable("user"); err != nil {
			t.Fatalf(err.Error())
		}
		t.Log("--------------- drop table pass -----------------")
	}
	t.Log("--------------- create Table OK -----------------")
}

func stopdb(cfg *SqlConfig, connect, dbname string) {
	cfg.CloseDB()
	DropDB(connect, dbname)
}
