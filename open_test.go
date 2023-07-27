package mysql

import "testing"

// mysqlの接続テスト
func TestMysqlOpen(t *testing.T) {
	t.Log("-------------- mysqlの接続テスト --------------")
	config := Setup("localhost", "3306", "mysql", "mysql", "database")
	t.Log("-------------- mysqlのOpen テスト --------------")
	err := config.Connect()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("-------------- mysqlのClose テスト --------------")
	err = config.Close()
	if err != nil {
		t.Fatal(err)
	}
}
