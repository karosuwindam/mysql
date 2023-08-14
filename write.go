package mysql

import (
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

// テーブル内にデータを書き込む
func (cfg *MySqlConfig) Add(tName string, data interface{}) error {
	//ポインター型でない場合はエラーを返す
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		return errors.New("data is not pointer")
	}
	//ポインター型の場合はポインターの中身を取得する
	v = v.Elem()
	switch v.Kind() {
	case reflect.Struct: //構造体の場合は構造体の中身を取得する
		return cfg.AddFromStruct(tName, data)
	case reflect.Slice: //スライスの場合はスライスの中身を取得する
		pv := reflect.ValueOf(v.Interface())

		for i := 0; i < pv.Len(); i++ {
			fi := pv.Index(i)
			fi = reflect.NewAt(fi.Type(), unsafe.Pointer(fi.UnsafeAddr()))
			f := fi.Interface()
			if err := cfg.AddFromStruct(tName, f); err != nil {
				return err
			}

		}
	}
	return nil
}

func (cfg *MySqlConfig) AddFromStruct(tName string, data interface{}) error {
	cmd, err := createAddCmd(tName, data)
	if err != nil {
		return err
	}
	_, err = cfg.db.Exec(cmd)
	return err
}

// 構造体からデータを挿入するSQL文を生成する
func createAddCmd(tName string, data interface{}) (string, error) {
	//dataがポインター型でない場合はエラーを返す
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		return "", errors.New("data is not pointer")
	}
	cmd := "INSERT INTO " + tName + " ("
	cmd2 := " VALUES ("
	//ポインターの中身を取得する
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return "", errors.New("data is not struct")
	}
	for i := 0; i < v.NumField(); i++ {

		//フィールド名を取得するもし、dbタグを持っている場合はそのタグを取得するまた、先頭が小文字の場合はスキップする
		name := v.Type().Field(i).Name
		if v.Type().Field(i).Tag.Get("db") != "" {
			name = v.Type().Field(i).Tag.Get("db")
		} else if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}

		if i != 0 && cmd2 != " VALUES (" {
			cmd += ", "
			cmd2 += ", "
		}

		cmd += name
		// 構造体の型を判定して、値を取得する
		switch v.Field(i).Kind() {
		case reflect.Int64, reflect.Int:
			num := v.Field(i).Int()
			cmd2 += strconv.Itoa(int(num))
		case reflect.String:
			cmd2 += "'" + v.Field(i).String() + "'"
		default:
			return "", errors.New("data type is not supported")
		}

	}
	cmd += ")"
	cmd2 += ")"
	return cmd + cmd2, nil
}
