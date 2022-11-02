package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type SqlConfig struct {
	db         *sql.DB // 開いたデータベースの値
	conectflag bool    //疎通フラグ
}

// CreateConectName()
//
// 接続用の文字列を作成
//
func CreateConectName(dbUser, dbPass, conectType, ipAddr, portN string) string {
	return dbUser + ":" + dbPass + "@" + conectType + "(" + ipAddr + ":" + portN + ")/"
}

// Setup(conect, dbname) = *SqlConfig, error
//
// セットアップ設定
// conect: 接続情報 user:password@tcp(ip:port)/
// dbname: 管理下のデータベース名
func Setup(conect, dbname string) (*SqlConfig, error) {
	tmp := &SqlConfig{}
	if db, err := openDB(conect); err != nil {
		return nil, err
	} else {
		str, err := listDB(db)
		if err != nil {
			return nil, err
		}
		if !str[dbname] {
			if err := createDB(conect, dbname); err != nil {
				return nil, err
			}
		}
		db.Close()
	}

	if db, err := openDB(conect + dbname); err != nil {
		return tmp, err
	} else {
		tmp.db = db
	}
	tmp.conectflag = true

	return tmp, nil
}

// (*SqlConfig)Ping
//
// 疎通確認
func (cfg *SqlConfig) Ping() error {
	return cfg.db.Ping()
}

//openDB()
//データベースを開く
func openDB(conect string) (*sql.DB, error) {
	db, err := sql.Open("mysql", conect+"?parseTime=true")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		fmt.Println(conect, "データベースの接続失敗")
		db.Close()
	}
	return db, err
}

// createDB(conect, dbname) = error
//
// conect: 接続の設定
// dbname: データベースの名前
func createDB(conect, dbname string) error {
	db, err := sql.Open("mysql", conect)
	if err != nil {
		fmt.Println(err.Error())
	}

	if err := db.Ping(); err != nil {
		fmt.Println(err.Error())
		return err
	}
	str := "CREATE DATABASE " + dbname + ";"
	_, err = db.Exec(str)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully created " + dbname + " database..")
	}
	db.Close()
	return err
}

// dropDB(conect, dbname) = error
//
// データベースの削除
func DropDB(conect, dbname string) error {
	db, err := sql.Open("mysql", conect)
	if err != nil {
		fmt.Println(err.Error())
	}

	if err := db.Ping(); err != nil {
		fmt.Println(err.Error())
		return err
	}
	str := "DROP DATABASE " + dbname + ";"
	_, err = db.Exec(str)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully Drop " + dbname + " database..")
	}
	db.Close()
	return err

}

// (*SqlConfig)listDB() = map[string]bool, error
//
// データベースの名前リストを取得
func listDB(db *sql.DB) (map[string]bool, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	out := map[string]bool{}
	cmd := "SHOW DATABASES;"
	rows, err := db.Query(cmd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmp := ""
		err := rows.Scan(&tmp)
		if err != nil {
			return nil, err
		}
		out[tmp] = true
	}
	return out, nil
}

// (*SqlConfig)CloseDB()
//
// dbを閉じる
func (cfg *SqlConfig) CloseDB() {
	cfg.conectflag = false
	cfg.db.Close()
}
