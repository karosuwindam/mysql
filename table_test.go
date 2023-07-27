package mysql

import (
	"fmt"
	"testing"
)

func TestCreateTable(t *testing.T) {
	cfg := Setup("localhost", "3306", "mysql", "mysql", "database")
	if err := cfg.Connect(); err != nil {
		return
	}
	defer cfg.Close()
	type TestTable struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	t.Log("-------------- mysqlのTable作成 テスト --------------")
	if err := cfg.CreateTableFromStruct("test_table", TestTable{}); err != nil {
		t.Fatal(err)
	}
	t.Log("-------------- mysqlのTable作成済み テスト --------------")
	if list, err := cfg.GetTableNames(); list[0] != "test_table" && err != nil {
		t.Fatal(list, err)
	}
	t.Log("-------------- mysqlのTable削除 テスト --------------")
	if err := cfg.DropTable("test_table"); err != nil {
		t.Fatal(err)
	}
	t.Log("-------------- mysqlのTable削除済み テスト --------------")
	if list, err := cfg.GetTableNames(); len(list) != 0 && err != nil {
		t.Fatal(list, err)
	}

}

// SQLコマンドの解析テスト
func TestParseCommand(t *testing.T) {
	cmd := "CREATE TABLE test_table (id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,name VARCHAR(255) NOT NULL,created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;"
	fmt.Println(parseCreateTableCommand(cmd))
}
