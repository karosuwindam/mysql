package mysql

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

// KeyWordOption 検索オプション
type KeyWordOption string

// 検索オプションの値
//
// AND keyword=data and
// OR keyword=data or
// AND_Like keyword like %keyword% and
// OR_LIKE keyword like %keyword% or
const (
	AND     KeyWordOption = "and"
	OR      KeyWordOption = "or"
	ANDLike KeyWordOption = "and_like"
	ORLike  KeyWordOption = "or_like"
)

// (*cfg)Read(tname, slice, v...) == error
//
// # MySQLからデータを読み取る
//
// tname(string):読み取り対象をテーブル名
// slice(*[]interface{}):読み取ったデータを格納する変数、ポインタ配列として入力
//
// v : map[string]stringまたは、keytypeの値を設定する
// v(map[string]string):検索対象のキーワード、空白は検索しない
// keytype(KeyWordOption):検索オプション (sqlite.AND or sqlite.OR or sqlite.ANDLike or sqlite.ORLike)
func (cfg *MySqlConfig) Read(tname string, slice interface{}, v ...interface{}) error {
	cmd, err := createReadCmd(tname, slice, v...)
	if err != nil {
		return err
	}
	rows, err := cfg.db.Query(cmd)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		data, err := sqlite3RowsReadData(slice)
		if err != nil {
			return err
		}
		err = rows.Scan(data...)
		if err != nil {
			return err
		}
		tmpdata, err1 := silceToMap(data, slice)
		if err1 != nil {
			return err1
		}
		mapToStruct(tmpdata, slice)
	}

	return err
}

// ReadToday 日間の更新分データ
func (cfg *MySqlConfig) ReadToday(tname string, slice interface{}, v ...interface{}) error {
	return cfg.readWhileTime(tname, slice, "today")
}

// ReadToWeek 週間の更新分データ
func (cfg *MySqlConfig) ReadToWeek(tname string, slice interface{}, v ...interface{}) error {
	return cfg.readWhileTime(tname, slice, "toweek")
}

// ReadToMonth 月間の更新分データ
func (cfg *MySqlConfig) ReadToMonth(tname string, slice interface{}, v ...interface{}) error {
	return cfg.readWhileTime(tname, slice, "tomonth")
}

// (*cfg)readWhileTime(tname, slice, v...) == error
//
// # MySQLから時間指定の更新のデータを読み取る
//
// tname(string):読み取り対象をテーブル名
// slice(*[]interface{}):読み取ったデータを格納する変数、ポインタ配列として入力
//
// v : 何を入力しても無効
func (cfg *MySqlConfig) readWhileTime(tname string, slice interface{}, v ...interface{}) error {
	cmd, err := createReadDayCmd(tname, slice, v...)
	if err != nil {
		return err
	}
	rows, err := cfg.db.Query(cmd)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		data, err := sqlite3RowsReadData(slice)
		if err != nil {
			return err
		}
		err = rows.Scan(data...)
		if err != nil {
			return err
		}
		tmpdata, err1 := silceToMap(data, slice)
		if err1 != nil {
			return err1
		}
		mapToStruct(tmpdata, slice)
	}

	return err
}

// createReadCmd(tname,stu,v...) = string,error
//
// 読み取り用のコマンドを作るコマンド
//
// tname(string) : 読み取り対象のテーブル
// slice(*[]interface{}) : 検索対象の指定用
// v : map[string]stringまたは、keytypeの値を設定する
// keyword(map[string]string) : 検索用のキーワードデータ
// keytype(KeyWordOption) : 検索オプション
func createReadCmd(tname string, slice interface{}, v ...interface{}) (string, error) {
	rt := reflect.TypeOf(slice)
	keyword := map[string]string{}
	keytype := AND
	ck := 0
	for _, data := range v {
		switch data.(type) {
		case map[string]string:
			if (ck & 0x1) > 0 {
				continue
			}
			keyword = data.(map[string]string)
			ck |= 0x1
		case KeyWordOption:
			if (ck & 0x2) > 0 {
				continue
			}
			keytype = data.(KeyWordOption)
			ck |= 0x2
		}
	}
	if rt.Kind() != reflect.Ptr {
		return "", errors.New("This input stu data is not pointer")
	}
	cmd := "SELECT * FROM" + " " + tname
	if len(keyword) == 0 {
		// return cmd, nil
	} else {
		cmd += " " + "WHERE" + " " + convertSerchCmd(slice, keyword, keytype)

	}

	return cmd, nil
}

// createReadDayCmd(tname,stu,v...) = string,error
//
// 日付指定の読み取り用のコマンドを作るコマンド
//
// tname(string) : 読み取り対象のテーブル
// slice(*[]interface{}) : 検索対象の指定用
// v(string) : today, toweek, tomonth
func createReadDayCmd(tname string, slice interface{}, v ...interface{}) (string, error) {
	if len(v) < 1 {
		return "", errors.New("input type err")
	}
	if reflect.TypeOf(v[0]).Kind() != reflect.String {
		return "", errors.New("input type err :" + reflect.TypeOf(v[0]).Kind().String())
	}
	cmd := "SELECT * FROM" + " " + tname
	nowtime := time.Now()
	switch v[0].(string) {
	case "today":
		cmd += " " + "WHERE " + "updated_at "
		cmd += "BETWEEN '" + nowtime.Format("2006-01-02") + "' AND '"
		cmd += nowtime.Add(24*time.Hour).Format("2006-01-02") + "'"
	case "toweek":
		cmd += " " + "WHERE " + "updated_at "
		cmd += "BETWEEN '" + nowtime.Add(-24*time.Hour*7).Format("2006-01-02") + "' AND '"
		cmd += nowtime.Add(24*time.Hour).Format("2006-01-02") + "'"
	case "tomonth":
		cmd += " " + "WHERE " + "updated_at "
		cmd += "BETWEEN '" + nowtime.Add(-24*time.Hour*30).Format("2006-01-02") + "' AND '"
		cmd += nowtime.Add(24*time.Hour).Format("2006-01-02") + "'"
	default:
		return "", errors.New("input err :" + v[0].(string))

	}
	return cmd, nil
}

// sqlite3RowsReadData(slice) = []interface{},error
//
// # MySQLから読み取ったデータを格納する変数を作る
//
// slice(*[]interface{}) : 変換もとになる構造体配列
func sqlite3RowsReadData(slice interface{}) ([]interface{}, error) {
	sv := reflect.ValueOf(slice)
	if sv.Type().Kind() != reflect.Ptr {
		fmt.Println(sv.Type().Kind())
		return []interface{}(nil), errors.New("Don't pointer data input data = " + sv.Type().Kind().String())
	}
	var output []interface{}
	tStruct := reflect.TypeOf(sv.Elem().Interface()).Elem()
	vStruct := reflect.New(tStruct)
	loaddata := vStruct.Elem().Interface()
	if loaddata == nil {
		return output, errors.New("Don't input tablemap data for" + string(""))
	}
	rt := reflect.TypeOf(loaddata)
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field((i))
		switch ft.Type.Kind() {
		case reflect.Int:
			i := int64(0)
			output = append(output, &i)
		case reflect.String:
			str := string("")
			output = append(output, &str)
		case reflect.Struct:
			// str := string("")
			var str time.Time
			output = append(output, &str)
		default:
			return output, errors.New("Error data type " + strconv.Itoa(int(ft.Type.Kind())))
		}
	}
	if len(output) != 0 {
		for i := 0; i < 2; i++ {
			var str time.Time
			output = append(output, &str)
		}

	}
	return output, nil
}

// silceToMap(silce, stu) = map[string]interface{},error
//
// 構造体とSQLから読み取った配列データから、map形式のデータに変換
//
// silce(*[]interface{}) : map名のベースとなる構造体
// stu([]interface{}) : map形式のデータの値
func silceToMap(silce []interface{}, stu interface{}) (map[string]interface{}, error) {
	output := map[string]interface{}{}
	if len(silce) == 0 {
		return output, nil
	}

	sv := reflect.ValueOf(stu)
	if sv.Type().Kind() != reflect.Ptr {
		return output, errors.New("Don't struct pointer input i=" + sv.Type().Kind().String())
	}
	ii := sv.Elem().Interface()
	tStruct := reflect.TypeOf(ii).Elem()
	vStruct := reflect.New(tStruct)
	ckStruct := reflect.TypeOf(vStruct.Elem().Interface())
	for i, data := range silce {
		dataf := reflect.ValueOf(data).Elem()
		var tmp interface{}
		if dataf.Kind() == timeKind {

			dataf = reflect.NewAt(dataf.Type(), unsafe.Pointer(dataf.UnsafeAddr())).Elem()
			tmp = dataf.Interface()
		} else {
			tmp = dataf.Interface()

		}
		if i < ckStruct.NumField() {
			f := ckStruct.Field(i)
			switch tmp.(type) {
			case int64:
				output[f.Name] = int(tmp.(int64))
			case int:
				output[f.Name] = tmp.(int)
			case string:
				output[f.Name] = tmp.(string)
			case time.Time:
				output[f.Name] = tmp.(time.Time)
			}
		} else {
			if i == len(silce)-2 {
				output["create_at"] = tmp.(time.Time)
			} else if i == len(silce)-1 {
				output["update_at"] = tmp.(time.Time)

			}
		}
	}
	return output, nil
}

// convertSerchCmd(silce,keyword, keytype) = string
//
// silce内の構造体に含まれるdbタグから検索用のコマンド内の値をを作成
//
// silce(*[]interface{}) : 検索コマンドベースとなるdbタグが含まれた構造体
// keyword(map[string]string) : 検索のdbタグのkey名とその値
// keytype : 検索コマンドを作るためのオプション
func convertSerchCmd(silce interface{}, keyword map[string]string, keytype KeyWordOption) string {
	cmd := ""
	sv := reflect.ValueOf(silce)
	if sv.Type().Kind() != reflect.Ptr {
		return ""

	}
	tStruct := reflect.TypeOf(sv.Elem().Interface()).Elem()
	vStruct := reflect.New(tStruct)
	ckStruct := reflect.TypeOf(vStruct.Elem().Interface())
	count := 0
	for i := 0; i < ckStruct.NumField(); i++ {
		f := ckStruct.Field(i)
		if keyword[f.Tag.Get("db")] != "" {
			if count != 0 {
				switch keytype {
				case ANDLike:
					cmd += " " + string(AND) + " "
				case ORLike:
					cmd += " " + string(OR) + " "
				default:
					cmd += " " + string(keytype) + " "
				}

			}
			if keytype == AND || keytype == OR {
				cmd += f.Tag.Get("db") + "="
				switch f.Type.Kind() {
				case reflect.Int:
					cmd += keyword[f.Tag.Get("db")]
				case reflect.String:
					cmd += "'" + keyword[f.Tag.Get("db")] + "'"
				}
			} else {
				cmd += f.Tag.Get("db") + " like "
				switch f.Type.Kind() {
				case reflect.Int:
					cmd += "'%" + keyword[f.Tag.Get("db")] + "%'"
				case reflect.String:
					cmd += "'%" + keyword[f.Tag.Get("db")] + "%'"
				}
			}
			count++
		}
	}

	return cmd
}
