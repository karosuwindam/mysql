package mysql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ifnot
//
// SQLコマンド作成時でＩＦ　EXISTSをつけるフラグ
type ifnot bool

const (
	ifnotOff ifnot = false
	ifnotOn  ifnot = true
)

// (*cfg)CreateTable(tname, stu) = error
//
// SQL内にテーブルを作成する
//
// tname(string) : 作成するテーブル名
// stu(interface{}) : 作成するテーブル内の構造体
func (cfg *SqlConfig) CreateTable(tname string, stu interface{}) error {
	var cmd string
	tlist, err := cfg.ReadTableList()
	if err != nil {
		return err
	}
	createflag := false
	for _, str := range tlist {
		if str == tname { //tableが作成済み
			createflag = true
			break
		}
	}
	if createflag { //tableが作成済み
		backcmd, err := cfg.ReadCreateTableCmd(tname)
		if err != nil {
			return err
		}
		cmd, err = createTableCmd(tname, stu, ifnotOff)
		if err != nil {
			return err
		}

		tNameA, dataA := createSqlTableAna(backcmd)
		tNameB, dataB := createSqlTableAna(cmd)
		if tNameA == tNameB && len(dataA) == len(dataB) { //登録コマンドと実行コマンドが同じ
			return nil

		} else { //登録コマンドと実行コマンドが異なる
			return nil

		}

	} else { //tableが作成していない
		cmd, err = createTableCmd(tname, stu, ifnotOn)
		if err != nil {
			return err
		}
		_, err = cfg.db.Exec(cmd)

	}

	return err

}

// (*cfg)CreateTable(tname, stu) = error
//
// SQL内にテーブルを更新する処理
//
// ToDo
// tname(string) : 作成するテーブル名
// stu(interface{}) : 作成するテーブル内の構造体
func (cfg *SqlConfig) UpdateTable(tname string, stu interface{}, slice interface{}) error {
	return nil
	var cmd string
	backcmd, err := cfg.ReadCreateTableCmd(tname)
	if err != nil {
		return err
	}
	if backcmd != "" { //tableが作成済み
		cmd, err = createTableCmd(tname, stu, ifnotOff)
		if err != nil {
			return err
		}

		tNameA, dataA := createSqlTableAna(backcmd)
		tNameB, dataB := createSqlTableAna(cmd)
		if tNameA == tNameB && len(dataA) == len(dataB) { //登録コマンドと実行コマンドが同じ
			return nil

		} else { //登録コマンドと実行コマンドが異なる
			return nil
		}

	} else { //tableが作成していない
		cmd, err = createTableCmd(tname, stu, ifnotOn)
		if err != nil {
			return err
		}
		_, err = cfg.db.Exec(cmd)

	}

	return err
}

// (*cfg)ReadTableList() = []string, error
//
// SQL内のテーブル名を取得
func (cfg *SqlConfig) ReadTableList() ([]string, error) {
	var output []string
	cmd, err := readTableAllCmd()
	if err != nil {
		return output, err
	}
	rows, err := cfg.db.Query(cmd)
	if err != nil {
		return output, err
	}
	defer rows.Close()
	for rows.Next() {
		str := ""
		err = rows.Scan(&str)
		if err != nil {
			return []string{}, err
		}
		output = append(output, str)
	}

	return output, err
}

// (*cfg)ReadCreateTableCmd(tname) = string, error
//
// SQL内の作成したテーブルのコマンド情報を取得
//
// tname(string) : 読み取り対象のテーブル
func (cfg *SqlConfig) ReadCreateTableCmd(tname string) (string, error) {
	var output string
	cmd, err := readCreateTableCmd(tname)
	if err != nil {
		return output, err
	}
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

	return output, err

}

// (*cfg)DropTable(tname) = error
//
// SQL内のテーブルを削除する
//
// tname(string) : 削除対象のテーブル
func (cfg *SqlConfig) DropTable(tname string) error {
	cmd, err := dropTableCmd(tname)
	if err != nil {
		return err
	}
	_, err = cfg.db.Exec(cmd)
	return err

}

// createTableCmd(tname, stu, flag) = string, error
//
// テーブルを作成するSQLコマンドを作成
// 構造体データにcreated_atとupdated_atを追加して、作成と更新のタイムスタンプをつける
//
// tname(string) : テーブル名
// stu(interface{}) : テーブル内のデータ精製オプション
// flag(ifnot) : IF NOT EXISTSをつけるオプション
func createTableCmd(tname string, stu interface{}, flag ifnot) (string, error) {
	cmd := "CREATE TABLE" + " "
	if flag == ifnotOn {
		cmd += "IF NOT EXISTS" + " "
	}
	if tname == "" {
		return "", errors.New("Don't input name data")
	}
	cmd += "" + tname + ""
	cmd += " ("
	if reflect.TypeOf(stu).Kind() != reflect.Struct {
		return "", errors.New("Don't input st data")
	}
	rt := reflect.TypeOf(stu)
	count := 0
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tmp := ""
		if i != 0 {
			cmd += ", "
		}
		switch f.Type.Kind() {
		case reflect.Int:
			tmp = f.Tag.Get("db")
			cmd += "" + tmp + " INT"
		case reflect.String:
			tmp = f.Tag.Get("db")
			cmd += "" + tmp + " VARCHAR(255)"
		case timeKind:
			tmp = f.Tag.Get("db")
			cmd += "" + tmp + " DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP"
		}
		if tmp == "id" {
			cmd += " NOT NULL AUTO_INCREMENT ,PRIMARY KEY (id)"
			count++
		} else if tmp == "" {
			return "", errors.New("Don't tag setup for " + f.Name)
		}
	}
	if count == 0 {
		return "", errors.New("Don't Struct data for \"id\"")
	}
	cmd += ", created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP"
	cmd += ", updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP"
	cmd += ")"
	return cmd, nil

}

// altertableCmd(cmdA, cmdB) = []string
//
// 作成コマンドを比較して追加作成用のテーブルを作るコマンド
// SQLiteではAlterコマンドは非対応のため未使用
//
// cmdA : 登録してあるコマンド
// cmdB : これから設定するコマンド
func altertableCmd(cmdA, cmdB string) []string {
	var output []string
	tnameA, dataA := createSqlTableAna(cmdA)
	tnameB, dataB := createSqlTableAna(cmdB)
	if tnameA != tnameB {
		return output
	}
	if len(dataA) < len(dataB) {
		bKey := ""
		for aKey, tdata := range dataB {
			if dataA[aKey] != tdata && bKey != "" {
				cmd := createAlterTableCmd(tnameB, bKey, aKey, tdata)
				output = append(output, cmd)
			}
			bKey = aKey

		}
	}

	return output
}

//updateTableAnabledCk(dataA,dataB) = bool
//
// mapデータを比較して、dataAの情報がdataBにすべて含まれていることを確認
// dataA(map[string]string) : 元のデータ
// dataB(map[string]string) : 切り替え先データ
func updateTableAnabledCk(dataA, dataB map[string]string) bool {
	if len(dataA) > len(dataB) {
		return false
	}
	for key, tData := range dataA {
		if dataB[key] != tData {
			return false
		}
	}
	return true
}

// createSqlTableAna(cmd) = string, map[string]string
//
// Table作成のSQLコマンドを解析して、テーブルとkey名と型のMapデータを作る
//
// cmd(string) : 解析用のコマンド
func createSqlTableAna(cmd string) (string, map[string]string) {
	tmp := strings.Split(cmd, "(")
	tmp1 := strings.Split(tmp[0], " ")
	if strings.ToLower(tmp1[1]) != "table" {
		return "", nil
	}
	tmp2 := strings.Split(tmp[1], ")")[0]
	tname := tmp1[len(tmp1)-1]
	if tmp1[len(tmp1)-1] == "" {
		tname = tmp1[len(tmp1)-2]
	}
	if tname[0] == "\""[0] {
		if tname[0] == tname[len(tname)-1] {
			tname = tname[1 : len(tname)-1]
		}
	}
	mdata := map[string]string{}
	for _, key := range strings.Split(tmp2, ",") {
		tmKey := strings.Split(key, " ")
		count := 0
		for ; count < len(tmKey)-1; count++ {
			if tmKey[count] != "" {
				break
			}
		}
		nKey := tmKey[count]
		if nKey[0] == "\""[0] {
			if nKey[0] == nKey[len(nKey)-1] {
				nKey = nKey[1 : len(nKey)-1]
			}
		}
		tKey := tmKey[count+1]
		mdata[nKey] = tKey
	}

	return tname, mdata

}

// createAlterTableCmd(tname,bKey,aKey,tdata) = string
//
// SQLite用のテーブルのカラム追加コマンドを作成
//
// tname(string) : 対象のテーブル
// bKey(string) : 挿入対象の前Key名
// aKey(string) : 挿入Key名
// tdata(string) : 挿入Keyに対応したKeyの型
func createAlterTableCmd(tname, bKey, aKey, tdata string) string {
	cmd := fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v AFTER %v", tname, aKey, tdata, bKey)
	return cmd
}

// dropTableCmd(tname) = string, error
//
// テーブルを削除するSQLコマンドを作る
//
// tname(string) : 削除対象のテーブル
func dropTableCmd(tname string) (string, error) {
	cmd := "DROP TABLE IF EXISTS" + " " + tname + ""
	return cmd, nil

}

// readTableAllCmd() = string, error
//
// SQLiteに登録してあるテーブルを取得するコマンドを作る
func readTableAllCmd() (string, error) {
	cmd := "SHOW TABLES"
	return cmd, nil

}

// readCreateTableCmd() = string, error
//
// SQLiteに登録してあるテーブルを作成したコマンドを読み取るSQLコマンドを作る
//
// tname(string) : 読み取り対象となるテーブル
func readCreateTableCmd(tname string) (string, error) {
	cmd := "SHOW CREATE TABLE " + tname + ""
	return cmd, nil
}
