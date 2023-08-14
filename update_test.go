package mysql_test

import (
	"mysql"
	"testing"
	"time"
)

// SQLのテーブル内のデータを更新する
func TestUpdate(t *testing.T) {
	type test struct {
		Id       int       `json:"id" db:"id"`
		Name     string    `json:"name" db:"name"`
		Num      int       `json:"num" db:"num"`
		timedata time.Time `json:"timedata" db:"timedata"`
	}
	cfg := mysql.Setup("localhost", "3306", "mysql", "mysql", "database")
	if err := cfg.Connect(); err != nil {
		return
	}
	defer cfg.Close()
	cfg.CreateTableFromStruct("testdelte", test{})
	defer cfg.DropTable("testdelte")
	var data test = test{0, "testdelte", 1, time.Now()}
	if err := cfg.Add("testdelte", &data); err != nil {
		t.Error(err)
	}
	data.Id = 1
	data.Name = "testupdate"
	timenow := time.Now()
	data.timedata = timenow
	t.Logf("Update data: %v", data)
	if err := cfg.Update("testdelte", &data); err != nil {
		t.Error(err)
	}
	var rdata []test = []test{}
	cfg.Read("testdelte", &rdata)
	if len(rdata) != 1 {
		t.Error("Update Error")
	}
	if rdata[0].Name != "testupdate" {
		t.Error("Update Error")
	}
	if rdata[0].timedata.Format(mysql.TimeLayout) != timenow.Format(mysql.TimeLayout) {
		t.Error("Update Error")
	}
}
