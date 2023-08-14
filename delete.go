package mysql

import "fmt"

// SQLのテーブルからid指定でデータを削除する
func (cfg *MySqlConfig) Delete(tName string, id int) error {
	cmd := fmt.Sprintf("delete from %s where id=%d", tName, id)
	_, err := cfg.db.Exec(cmd)
	return err
}
