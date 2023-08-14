package mysql

import (
	"math/rand"
	"strings"
	"testing"
)

// SQLのコマンド作成テスト
func TestSqlReadCmd(t *testing.T) {
	type Test struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		Age     int    `db:"age"`
		KeyWord string `db:"key"`
	}
	if cmd, err := createReadCmd("test", &[]Test{}); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
	if cmd, err := createReadCmd("test", &[]Test{}, map[string]string{"name": "test1"}); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
	if cmd, err := createReadCmd("test", &[]Test{}, map[string]string{"name": "test1", "age": "20"}, ORLike); err != nil {
		t.Error(err)
	} else {
		t.Log(cmd)
	}
	if cmd, err := createReadCmd("text", &[]Test{}, map[string]string{"keyword": "ag"}, ORLike); err != nil {
		t.Log(err)
	} else {
		t.Errorf("Not err output cmd:%s", cmd)
	}
}

// データを読み込むテスト
func TestRead(t *testing.T) {
	type Sample struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		Age     int    `db:"age"`
		KeyWord string `db:"keyword"`
	}
	cfg := Setup("localhost", "3306", "mysql", "mysql", "database")
	if err := cfg.Connect(); err != nil {
		return
	}
	defer cfg.Close()
	cfg.CreateTableFromStruct("test", Sample{})

	var data []Sample = []Sample{
		Sample{0, "test1", 20, ""},
		Sample{0, "test2", 30, ""},
		Sample{0, "test3", 40, ""},
		Sample{0, "test4", 50, ""},
		Sample{0, "test5", 60, ""},
		Sample{0, "test6", 70, ""},
	}
	strst := []string{}
	for i := 0; i < len(data); i++ {
		strst = append(strst, randomString(10))
	}
	for i := 0; i < len(data); i++ {
		data[i].KeyWord = strst[i]
	}
	if err := cfg.Add("test", &data); err != nil {
		t.Error(err)
	}
	t.Log("Chack Read All Data")
	var readData []Sample = []Sample{}
	if err := cfg.Read("test", &readData); err != nil {
		t.Error(err)
	}
	for i := 0; i < len(data); i++ {
		if data[i].Name != readData[i].Name {
			t.Errorf("Not match data:%s != %s", data[i].Name, readData[i].Name)
		}
		if data[i].Age != readData[i].Age {
			t.Errorf("Not match data:%d != %d", data[i].Age, readData[i].Age)
		}
		if data[i].KeyWord != readData[i].KeyWord {
			t.Errorf("Not match data:%s != %s", data[i].KeyWord, readData[i].KeyWord)
		}
	}
	s := randomString(3)
	for {
		flag := false
		//作成文字列がデータに含まれているかチェック
		for _, tstr := range strst {
			if strings.Index(tstr, s) != -1 {
				flag = true
				break
			}
		}
		if flag {
			break
		}
		s = randomString(3)
	}
	t.Logf("Chack Read Data Where KeyWord Like %s", s)
	readData = []Sample{}
	if err := cfg.Read("test", &readData, map[string]string{"keyword": s}, ORLike); err != nil {
		t.Error(err)
	}
	for _, d := range readData {
		flag := false
		for i := 0; i < len(strst); i++ {
			if data[i].Name == d.Name && data[i].Age == d.Age && data[i].KeyWord == d.KeyWord {
				flag = true
				break
			}
		}
		if !flag {
			t.Errorf("Not match data:%s", d.Name)
		}
	}
	t.Log("Chack Read Data Where Name = test1")

	//テーブル削除
	if err := cfg.DropTable("test"); err != nil {
		t.Error(err)
	}

}

// ランダムで英数字の文字の文字列を生成する
func randomString(s int) string {
	var str string
	for i := 0; i < s; i++ {
		str += string(rune('a' + rand.Intn(26)))
	}
	return str

}
