package mysql

import (
	"errors"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

// SQLのデータを書き換える
func (cfg *MySqlConfig) Update(tName string, data interface{}) error {
	cmd, err := createUpdateCmd(tName, data)
	if err != nil {
		return err
	}
	_, err = cfg.db.Exec(cmd)
	return err
}

// SQLのデータを書き換えるコマンドを作成する
func createUpdateCmd(tName string, data interface{}) (string, error) {
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return "", errors.New("data is not pointer")
	}
	//data内の構造体からSQLのUPDATE文を作成する
	id := 0
	cmd := "update " + tName + " set "
	cmd2 := ""
	v := reflect.ValueOf(data).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		//フィールド名を取得する
		name := t.Field(i).Name
		if tag := t.Field(i).Tag.Get("db"); tag != "" {
			name = tag
		} else if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}
		if name == "id" || name == "Id" {
			id = int(f.Int())
			continue
		}

		//フィールドの値を取得する
		if f.Kind() == timeKind {
			fi := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr()))
			f = fi.Elem()
		}
		value := f.Interface()
		if value == nil {
			continue
		}
		if cmd2 != "" {
			cmd2 += ","
		}
		switch value.(type) {
		case string:
			cmd2 += name + "='" + value.(string) + "'"
		case int:
			cmd2 += name + "=" + strconv.Itoa(value.(int))
		case int64:
			cmd2 += name + "=" + strconv.FormatInt(value.(int64), 10)
		case float64:
			cmd2 += name + "=" + strconv.FormatFloat(value.(float64), 'f', -1, 64)
		case bool:
			if value.(bool) {
				cmd2 += name + "=1"
			} else {
				cmd2 += name + "=0"
			}
		case time.Time:
			cmd2 += name + "='" + value.(time.Time).Format(TimeLayout) + "'"
		default:
			return "", errors.New("data type is not match")
		}
	}
	if cmd2 == "" {
		return "", errors.New("data is not match")
	}
	now := time.Now().Format(TimeLayout)
	cmd += cmd2 + ",updated_at='" + now + "' where id=" + strconv.Itoa(id)

	return cmd, nil

}
