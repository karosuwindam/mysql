package mysql

import (
	"fmt"
	"testing"
)

// SQLコマンドを作成するテスト
func TestCreateSQLWrite(t *testing.T) {
	type test struct {
		Name string `json:"name" db:"name"`
		Num  int    `json:"num" db:"num"`
		auth int
	}
	var data test = test{"test", 1, 0}
	cmd, err := createAddCmd("test", &data)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(cmd)

}

// データを書き込むテスト
func TestWrite(t *testing.T) {

	type Sample struct {
		Id   int    `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
		Num  int    `json:"num" db:"num"`
		auth int
	}
	var data Sample = Sample{0, "test", 1, 0}
	cfg := Setup("localhost", "3306", "mysql", "mysql", "database")
	if err := cfg.Connect(); err != nil {
		return
	}
	defer cfg.Close()
	cfg.CreateTableFromStruct("test", Sample{})
	if err := cfg.Add("test", &data); err != nil {
		t.Error(err)
	}
	var sData []Sample = []Sample{
		Sample{0, "test1", 1, 0},
		Sample{0, "test2", 2, 0},
	}
	if err := cfg.Add("test", &sData); err != nil {
		t.Error(err)
	}
	//テーブル削除
	cfg.DropTable("test")

}
