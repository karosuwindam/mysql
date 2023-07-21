package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// MySqlConfig Mysqlのテーブル設定
type MySqlConfig struct {
	Host     string  // ホスト名
	Port     string  // ポート番号
	Username string  // ユーザー名
	Password string  // パスワード
	Database string  // データベース名
	db       *sql.DB // DB接続情報
}

// Setup Mysqlの設定
//
// 基本セットアップ
//
// host: ホスト名
// port: ポート番号
// username: ユーザー名
// password: パスワード
// database: データベース名
func Setup(host string, port string, username string, password string, database string) MySqlConfig {
	return MySqlConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}
}

// Mysqlの接続情報を取得する
func (config *MySqlConfig) GetConnectInfo() string {
	return config.Username + ":" + config.Password + "@tcp(" + config.Host + ":" + config.Port + ")/" + config.Database + "?parseTime=true"
}

// Mysqlへ接続する
func (config *MySqlConfig) Connect() error {
	db, err := sql.Open("mysql", config.GetConnectInfo())
	if err != nil {
		return err
	}
	config.db = db
	return nil
}

// Mysqlへの接続を切断する
func (config *MySqlConfig) Close() error {
	return config.db.Close()
}
