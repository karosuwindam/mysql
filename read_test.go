package mysql

import "testing"

// SQLのコマンド作成テスト
func TestSqlReadCmd(t *testing.T) {
	type Test struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		Age     int    `db:"age"`
		KeyWord string `db:"key"`
	}
	// var test []Test = []Test{
	// 	Test{Id: 1, Name: "test1", Age: 20, KeyWord: "ageag"},
	// 	Test{Id: 2, Name: "test2", Age: 30, KeyWord: "agbg"},
	// 	Test{Id: 3, Name: "test3", Age: 40, KeyWord: "herh"},
	// 	Test{Id: 4, Name: "test4", Age: 50, KeyWord: "hrshs"},
	// }
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
}
