package mysql

import "testing"

// SQLのテーブルからid指定でデータ削除のテスト
func TestSQLDelete(t *testing.T) {
	type test struct {
		Id   int    `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
		Num  int    `json:"num" db:"num"`
		auth int
	}
	cfg := Setup("localhost", "3306", "mysql", "mysql", "database")
	if err := cfg.Connect(); err != nil {
		return
	}
	defer cfg.Close()
	cfg.CreateTableFromStruct("test", test{})
	var data test = test{0, "test", 1, 0}
	if err := cfg.Add("test", &data); err != nil {
		t.Error(err)
	}
	if err := cfg.Delete("test", 1); err != nil {
		t.Error(err)
	}
	var rdata []test = []test{}
	cfg.Read("test", &rdata)
	if len(rdata) != 0 {
		t.Error("Delete Error")
	}
	var wdata []test = []test{
		test{0, "test1", 1, 0},
		test{0, "test2", 2, 0},
	}
	if err := cfg.Add("test", &wdata); err != nil {
		t.Error(err)
	}
	rdata = []test{}
	cfg.Read("test", &rdata)
	if err := cfg.Delete("test", rdata[0].Id); err != nil {
		t.Error(err)
	}
	rdata = []test{}
	cfg.Read("test", &rdata)
	if len(rdata) != 1 {
		t.Error("Delete Error")
	} else if rdata[0].Name != "test2" {
		t.Error("Delete Error")
	}
	//テーブル削除
	cfg.DropTable("test")
}
