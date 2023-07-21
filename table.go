package mysql

import (
	"fmt"
	"reflect"
	"strings"
)

// SQL内に構造体情報からテーブルを作成する
func (cfg *MySqlConfig) CreateTableFromStruct(tname string, table interface{}) error {
	var cmd string = ""
	var err error
	//SQL内でテーブル名で検索して、テーブルが存在するか確認する
	createCmd, _ := cfg.GetTableCommand(tname)
	if createCmd == "" { //テーブルが存在しない場合
		cmd, err = createCreateTableCommand(tname, table)
		if err != nil {
			return err
		}
		_, err = cfg.db.Exec(cmd)
	}

	//構造体からテーブルを作成するコマンドを作る

	return err
}

// SQL内にテーブルを更新する処理

// SQL内のテーブル名のリストを取得
func (cfg *MySqlConfig) GetTableNames() ([]string, error) {
	var output []string = nil
	// SQL内のテーブル名を取得するコマンドを作る
	cmd, err := createGetTableNamesCommand()
	if err != nil {
		return output, err
	}
	// SQL内のテーブル名を取得する
	rows, err := cfg.db.Query(cmd)
	if err != nil {
		return output, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return output, err
		}
		output = append(output, name)
	}

	return output, nil
}

// SQL内の作成したテーブルのコマンド情報を取得
func (cfg *MySqlConfig) GetTableCommand(tname string) (string, error) {
	var output string = ""
	// SQL内のテーブル名を取得するコマンドを作る
	cmd, err := createGetTableCreateCommand(tname)
	if err != nil {
		return output, err
	}
	// SQL内のテーブル名を取得する
	rows, err := cfg.db.Query(cmd)
	if err != nil {
		return output, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name, &output)
		if err != nil {
			return "", err
		}
	}

	return output, nil
}

// SQL内のテーブルを削除する
func (cfg *MySqlConfig) DropTable(tname string) error {
	// SQL内のテーブル名を取得するコマンドを作る
	cmd, _ := createDropTableCommand(tname)
	// SQL内のテーブル名を取得する
	_, err := cfg.db.Query(cmd)
	if err != nil {
		return err
	}

	return nil
}

// 構造体からテーブルを作成するコマンドを作る
func createCreateTableCommand(tname string, table interface{}) (string, error) {
	cmd := ""
	idFlag := false
	if tname == "" {
		return cmd, fmt.Errorf("テーブル名が指定されていません")
	}
	if reflect.TypeOf(table).Kind() != reflect.Struct {
		return cmd, fmt.Errorf("構造体ではありません")
	}
	cmd = "CREATE TABLE " + tname + " ("
	//構造体のフィールドを取得する
	t := reflect.TypeOf(table)
	for i := 0; i < t.NumField(); i++ {
		//フィールド名を取得する
		f := t.Field(i)
		//フィールド名を取得するもし、dbタグを持っている場合はそのタグを取得する
		name := f.Name
		if f.Tag.Get("db") != "" {
			name = f.Tag.Get("db")
		}

		//フィールドの型を取得する
		kind := f.Type.Kind()
		//フィールドのタグを取得する
		tag := f.Tag
		//フィールドのタグからSQLのカラム名を取得する
		column := tag.Get("column")
		if column == "" {
			column = name
		}
		//フィールドのタグからSQLのカラムの型を取得する
		columnType := tag.Get("type")
		if columnType == "" {
			switch kind {
			case reflect.String:
				columnType = "VARCHAR(255)"
			case reflect.Int:
				columnType = "INT"
			case reflect.Int64:
				columnType = "BIGINT"
			case reflect.Float32:
				columnType = "FLOAT"
			case reflect.Float64:
				columnType = "DOUBLE"
			case reflect.Bool:
				columnType = "BOOLEAN"
			default:
				return cmd, fmt.Errorf("未対応の型です")
			}
		}

		//フィールドのタグからSQLのカラムのオプションを取得する
		columnOption := tag.Get("option")
		if columnOption == "" {
			columnOption = "NOT NULL"
		}
		// コマンドを作成する
		if strings.ToLower(column) == "id" { //idの場合は自動でプライマリーキーを設定する
			cmd += column + " " + columnType + " " + "PRIMARY KEY AUTO_INCREMENT NOT NULL,"
			idFlag = true
		} else if column != "" {
			cmd += column + " " + columnType + " " + columnOption + ","
		} else {
			return cmd, fmt.Errorf("カラム名が指定されていません")
		}
	}
	if !idFlag {
		return cmd, fmt.Errorf("idが指定されていません")
	}
	// 作成と更新タイムスタンプ情報を追加する
	cmd += "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,"
	cmd += "updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
	cmd += ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;"

	return cmd, nil
}

// Table作成のSQLコマンドを解析して、テーブル名とカラム名とカラムの型のMapを作成する
func parseCreateTableCommand(cmd string) (string, map[string]string, error) {
	var tname string = ""
	var output map[string]string = nil
	// テーブル名を取得する
	tname = strings.Split(cmd, " ")[2]
	// カラム名とカラムの型を取得する
	output = make(map[string]string)
	//最初で"("が部分移行を取得する
	if i := strings.Index(cmd, "("); i >= 0 {
		cmd = cmd[i+1:]
	}
	//最後で")"が最後にを取得する

	//カンマで分割する
	for _, line := range strings.Split(cmd, ",") {
		//カラム名とカラムの型を取得する
		column := strings.Split(line, " ")[0]
		columnType := strings.Split(line, " ")[1]
		output[column] = columnType
	}

	// カラム名とカラムの型を取得する
	for _, line := range strings.Split(cmd, "\n") {
		if strings.Contains(line, "PRIMARY KEY") {
			break
		}
		if strings.Contains(line, "KEY") {
			continue
		}
		if strings.Contains(line, "CONSTRAINT") {
			continue
		}
		if strings.Contains(line, "FOREIGN KEY") {
			continue
		}
		if strings.Contains(line, "UNIQUE KEY") {
			continue
		}
		if strings.Contains(line, "PRIMARY KEY") {
			continue
		}
	}
	return tname, output, nil
}

// テーブルを削除するSQLコマンドを作る
func createDropTableCommand(tname string) (string, error) {
	cmd := "DROP TABLE IF EXISTS " + tname
	return cmd, nil
}

// SQL内のテーブル名を取得するコマンドを作る
func createGetTableNamesCommand() (string, error) {
	cmd := "SHOW TABLES"
	return cmd, nil
}

// SQL内に登録してあるテーブルを作成したコマンドを読み取るSQLコマンドを作る
func createGetTableCreateCommand(tname string) (string, error) {
	cmd := "SHOW CREATE TABLE " + tname
	return cmd, nil
}
