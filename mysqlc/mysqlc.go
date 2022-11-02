package mysqlc

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//SqlConfigはデータベースに関する配列
type SqlConfig struct {
	dataSourceName string
	dbname         string
	setupflag      bool
	dbpath         *sql.DB
	tableName      string
}

var column_name []string
var column_type []string

//sqlSetup はデータベースの基本設定
func (t *SqlConfig) SqlSetup(dbUser, dbPass, conectType, ipAddr, portN, dbName string) error {
	t.dataSourceName = dbUser + ":" + dbPass + "@" + conectType + "(" + ipAddr + ":" + portN + ")/"
	t.dbname = dbName
	t.setupflag = true
	db, err := t.openDB()
	if err != nil {
		err = t.createDB()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		db, err = t.openDB()
	}
	t.dbpath = db
	return err
}

//TableSetupDB()
func (t *SqlConfig) TableSetupDB(table string, columnName, columnType []string) {
	t.tableName = table
	column_name = columnName
	column_type = columnType

}

//createDB はデータベースを作成する
func (t *SqlConfig) createDB() error {
	if !t.setupflag {
		return errors.New("not run setup")
	}
	db, err := sql.Open("mysql", t.dataSourceName)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	str := "CREATE DATABASE " + t.dbname + ";"
	_, err = db.Exec(str)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully created database..")
	}
	db.Close()
	return err
}

//openDB()
//データベースを開く
func (t *SqlConfig) openDB() (*sql.DB, error) {
	if !t.setupflag {
		return nil, errors.New("not run setup")
	}
	db, err := sql.Open("mysql", t.dataSourceName+t.dbname+"?parseTime=true")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("データベースの接続失敗")
		db.Close()
	} else {
		fmt.Println("データベースの接続成功")
	}
	return db, err
}

//CloseDB()
//
func (t *SqlConfig) CloseDB() error {
	t.setupflag = false
	err := t.dbpath.Close()
	return err
}

//CreateTableDB()
func (t *SqlConfig) CreateTableDB() {
	if !t.setupflag {
		return
	}
	cmd := "create table "
	cmd += t.tableName + " "
	cmd += "("
	cmd += "id" + " " + "int " + "NOT NULL AUTO_INCREMENT" + ","
	for i := 0; i < len(column_type); i++ {
		cmd += column_name[i] + " " + column_type[i] + ","
	}
	// cmd += "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,"
	cmd += "created_at DATETIME NOT NULL,"
	// cmd += "updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,"
	cmd += "updated_at DATETIME NOT NULL,"
	cmd += "PRIMARY KEY (id)"
	cmd += ")"
	fmt.Println(cmd)
	stmt, err := t.dbpath.Prepare(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("データベース作成")
	}

}

//AddDB()
//
func (t *SqlConfig) AddDB(v ...interface{}) {
	if !t.setupflag {
		return
	}
	time_now := time.Now()
	cmd := "insert into " + t.tableName + "("
	for i := 0; i < len(column_name); i++ {
		if i == 0 {
			cmd += column_name[i]
		} else {
			cmd += "," + column_name[i]
		}
	}
	cmd += "," + "created_at"
	cmd += "," + "updated_at"
	cmd += ") values("
	for i := 0; i < len(column_name); i++ {
		if i == 0 {
			cmd += "?"
		} else {
			cmd += ",?"
		}
	}
	cmd += ",'" + time_now.Format("2006-01-02 15:04:05.999999999") + "'"
	cmd += ",'" + time_now.Format("2006-01-02 15:04:05.999999999") + "'"
	// cmd += "created_at='" + time_now.Format("2006-01-02 15:04:05.999999999") + "',"
	// cmd += "updated_at=`" + time_now.Format("2006-01-02 15:04:05.999999999") + "`,"
	cmd += ")"
	fmt.Println(cmd)
	stmt, err := t.dbpath.Prepare(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(v...)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("追加成功")
	}
}

//ReadAllDB()
func (t *SqlConfig) ReadAllDB() (*sql.Rows, error) {
	cmd := "select * from " + t.tableName
	row, err := t.dbpath.Query(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	return row, err
}

//SerchTimeDB()
func (t *SqlConfig) SerchTimeDB(num int) (*sql.Rows, error) {
	nowtime := time.Now()
	cmd := "select * from " + t.tableName
	switch num {
	case 1: //week
		cmd += " " + "where " + "updated_at "
		cmd += "between '" + nowtime.Add(-24*time.Hour*7).Format("2006-01-02") + "' and '"
		cmd += nowtime.Add(24*time.Hour).Format("2006-01-02") + "'"
	case 2: //month
		cmd += " " + "where " + "updated_at "
		cmd += "between '" + nowtime.Add(-24*time.Hour*30).Format("2006-01-02") + "' and '"
		cmd += nowtime.Add(24*time.Hour).Format("2006-01-02") + "'"
	default: //today
		cmd += " " + "where " + "updated_at "
		cmd += "between '" + nowtime.Format("2006-01-02") + "' and '"
		cmd += nowtime.Add(24*time.Hour).Format("2006-01-02") + "'"
	}
	row, err := t.dbpath.Query(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	return row, err
}

//ReadIdDB()
func (t *SqlConfig) ReadIdDB(Id string) (*sql.Rows, error) {
	cmd := "select * from " + t.tableName
	cmd += " " + "where id=" + Id
	row, err := t.dbpath.Query(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	return row, err
}

func (t *SqlConfig) SeachReadDB(word string, serchKey []string) (*sql.Rows, error) {
	cmd := "select * from " + t.tableName
	cmd += " " + "where "
	for i := 0; i < len(serchKey); i++ {
		if i == 0 {
			cmd += serchKey[i] + " " + "like '%" + word + "%'"
		} else {
			cmd += " or " + serchKey[i] + " " + "like '%" + word + "%'"
		}
	}
	row, err := t.dbpath.Query(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	return row, err
}

//EditDB()
//
func (t *SqlConfig) EditDB(No string, v ...interface{}) {
	if !t.setupflag {
		return
	}
	time_now := time.Now()
	cmd := "update " + t.tableName + " set "
	for i := 0; i < len(column_name); i++ {
		if i == 0 {
			cmd += column_name[i] + "=?"
		} else {
			cmd += "," + column_name[i] + "=?"
		}
	}
	cmd += ",updated_at='" + time_now.Format("2006-01-02 15:04:05.999999999") + "'"
	cmd += " where id=" + No
	fmt.Println(cmd)
	stmt, err := t.dbpath.Prepare(cmd)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(v...)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("編集成功")
	}
}

//DeleteDB()
//
func (t *SqlConfig) DeleteDB(No int) {
	if !t.setupflag {
		return
	}
	cmd := "delete from " + t.tableName + " where id=?"
	stmtDelete, err := t.dbpath.Prepare(cmd)
	if err != nil {
		panic(err.Error())
	}
	defer stmtDelete.Close()

	result, err := stmtDelete.Exec(No)
	if err != nil {
		panic(err.Error())
	}

	_, err = result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

}
