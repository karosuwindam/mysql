package mysql

import "testing"

func TestOpenList(t *testing.T) {
	t.Logf(CreateConectName("root", "root", "tcp", "localhost", "3306"))
	if db, err := openDB(CreateConectName("root", "root", "tcp", "localhost", "3306")); err != nil {
		t.Fatalf(err.Error())
	} else {
		defer db.Close()
		if str, err := listDB(db); err != nil {
			t.Fatalf(err.Error())
		} else {
			t.Log(str)
		}
	}
}

func TestSetup(t *testing.T) {
	connect := CreateConectName("root", "root", "tcp", "localhost", "3306")
	if cfg, err := Setup(connect, "db_write"); err != nil {
		t.Fatalf(err.Error())
	} else {
		if str, err := listDB(cfg.db); err != nil {

		} else {
			if !str["db_write"] {
				t.FailNow()
			}
		}
		cfg.CloseDB()
	}
	if err := DropDB(connect, "db_write"); err != nil {
		t.Fatalf(err.Error())
	}
}
