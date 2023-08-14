package mysql

import (
	"database/sql"
	"errors"
	"reflect"
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

// SQLからデータを読み込み構造体に格納するなお、map[string]stringで上限を絞ることができる
func (cfg *MySqlConfig) Read(tName string, slice interface{}, v ...interface{}) error {
	//ポインター型でない場合はエラーを返す
	sl := reflect.ValueOf(slice)
	if sl.Kind() != reflect.Ptr {
		return errors.New("data is not pointer")
	}
	cmd, err := createReadCmd(tName, slice, v...)
	if err != nil {
		return err
	}
	rows, err := cfg.db.Query(cmd)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		data, err := createDataFromRows(slice)
		if err != nil {
			return err
		}
		err = rows.Scan(data...)
		if err != nil {
			return err
		}
		//読み取ったデータをmap[string]stringに変換する
		m, err := createMapFromData(data, slice)
		if err != nil {
			return err
		}
		//map[string]stringのデータを構造体に変換する
		err = createStructFromMap(m, slice)
		if err != nil {
			return err
		}

	}
	return nil
}

// map[string]stringからSQLのデータを読み込むコマンドを作成する
func createReadCmd(tName string, slice interface{}, v ...interface{}) (string, error) {
	keyword := map[string]string{}
	keytype := AND
	for _, data := range v {
		switch data.(type) {
		case map[string]string:
			for key, value := range data.(map[string]string) {
				keyword[key] = value
			}
		case KeyWordOption:
			keytype = data.(KeyWordOption)
		}
	}
	cmd := "SELECT * FROM " + tName
	if len(keyword) > 0 {
		if cmd2 := createWhereCmd(slice, keyword, keytype); cmd2 != "" {
			cmd += " WHERE " + cmd2
		} else {
			return "", errors.New("keyword is not match")
		}
	}

	return cmd, nil
}

// WHEREより後ろのSQL文を作成する
func createWhereCmd(slice interface{}, keyword map[string]string, keytype KeyWordOption) string {
	//ポインター型でない場合はエラーを返す
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Ptr {
		return ""
	}
	//ポインターの中身を取得する
	if v.Elem().Kind() != reflect.Slice {
		return ""
	}
	//ポインタの中身から新しく構造体の型を作る
	vs := reflect.New(reflect.TypeOf(v.Elem().Interface()).Elem())
	vv := reflect.TypeOf(vs.Elem().Interface())

	cmd := ""
	//keyword内のデータで構造体のフィールド名と一致するものを取得する
	for i := 0; i < vv.NumField(); i++ {
		f := vv.Field(i)
		//フィールド名を取得するもし、dbタグを持っている場合はそのタグを取得するまた、先頭が小文字の場合はスキップする
		name := f.Name
		if tag := f.Tag.Get("db"); tag != "" {
			name = tag
		} else if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}
		//keyword内にnameが存在するか確認する
		if _, ok := keyword[name]; ok {
			if cmd != "" {
				if keytype == ANDLike || keytype == AND {
					cmd += " AND "
				} else {
					cmd += " OR "
				}
			}
			//keyword内のnameの値を取得する
			value := keyword[name]
			//keyword内のnameの値を削除する
			delete(keyword, name)
			if keytype == ANDLike || keytype == ORLike {
				cmd += name + " LIKE '%" + value + "%' "
			} else {
				cmd += name + "='" + value + "' "
			}
		}
	}
	return cmd
}

// Mysqlから読み取ったデータを格納する構造体の変数を作成する
func createDataFromRows(slice interface{}) ([]interface{}, error) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Ptr {
		return nil, errors.New("data is not pointer")
	}
	//ポインターの中身を取得する
	if sv.Elem().Kind() != reflect.Slice {
		return nil, errors.New("data is not slice")
	}
	var output []interface{}
	tStruct := reflect.TypeOf(sv.Elem().Interface()).Elem()
	vStruct := reflect.New(tStruct)
	if vStruct.Elem().Interface() == nil {
		return nil, errors.New("data is nil")
	}
	rt := reflect.TypeOf(vStruct.Elem().Interface())
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		name := ft.Name
		if tag := ft.Tag.Get("db"); tag != "" {
			name = tag
		} else if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}
		switch ft.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i := sql.NullInt64{}
			output = append(output, &i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i := sql.NullInt64{}
			output = append(output, &i)
		case reflect.Float32, reflect.Float64:
			i := sql.NullFloat64{}
			output = append(output, &i)
		case reflect.String:
			i := sql.NullString{}
			output = append(output, &i)
		case reflect.Bool:
			i := sql.NullBool{}
			output = append(output, &i)
		default:
			return nil, errors.New("data type is not match")
		}
	}
	if len(output) != 0 {
		//2つの時間変数を追加する
		for i := 0; i < 2; i++ {
			i := sql.NullTime{}
			output = append(output, &i)
		}
	}
	return output, nil
}

// Mysqlから読み取ったデータをmap[string]stringに変換する
func createMapFromData(silce []interface{}, stu interface{}) (map[string]interface{}, error) {
	output := map[string]interface{}{}
	if len(silce) == 0 {
		return output, nil
	}
	sv := reflect.ValueOf(stu)
	if sv.Kind() != reflect.Ptr {
		return nil, errors.New("data is not pointer")
	}
	//ポインターの中身を取得する
	if sv.Elem().Kind() != reflect.Slice {
		return nil, errors.New("data is not slice")
	}
	ii := sv.Elem().Interface()
	tStruct := reflect.TypeOf(ii).Elem()
	vStruct := reflect.New(tStruct)
	ckStruct := reflect.TypeOf(vStruct.Elem().Interface())
	if vStruct.Elem().Interface() == nil {
		return nil, errors.New("data is nil")
	}
	for i, data := range silce {
		df := reflect.ValueOf(data).Elem()
		var tmp interface{}
		if df.Kind() == timeKind || df.Kind() == reflect.TypeOf(sql.NullTime{}).Kind() {
			df = reflect.NewAt(df.Type(), unsafe.Pointer(df.UnsafeAddr())).Elem()
			tmp = df.Interface()

		} else {
			tmp = df.Interface()
		}
		if i < ckStruct.NumField() {
			ft := ckStruct.Field(i)
			switch tmp.(type) {
			case int64:
				output[ft.Name] = tmp.(int64)
			case float64:
				output[ft.Name] = tmp.(float64)
			case string:
				output[ft.Name] = tmp.(string)
			case bool:
				output[ft.Name] = tmp.(bool)
			case time.Time:
				output[ft.Name] = tmp.(time.Time)
			case sql.NullInt64:
				output[ft.Name] = tmp.(sql.NullInt64).Int64
			case sql.NullFloat64:
				output[ft.Name] = tmp.(sql.NullFloat64).Float64
			case sql.NullString:
				output[ft.Name] = tmp.(sql.NullString).String
			case sql.NullBool:
				output[ft.Name] = tmp.(sql.NullBool).Bool
			case sql.NullTime:
				output[ft.Name] = tmp.(sql.NullTime).Time
			default:
				return nil, errors.New("data type is not match")
			}

		} else {
			switch tmp.(type) {
			case time.Time:
				tmp = tmp.(time.Time)
			case sql.NullTime:
				tmp = tmp.(sql.NullTime).Time
			default:
				return nil, errors.New("data type is not match")
			}

			if i == len(silce)-2 {
				output["create_at"] = tmp
			} else if i == len(silce)-1 {
				output["updata_at"] = tmp
			}
		}
	}
	return output, nil
}

// map[string]interface{}を構造体に変換する
func createStructFromMap(data map[string]interface{}, stu interface{}) error {
	sv := reflect.ValueOf(stu)
	if sv.Kind() != reflect.Ptr {
		return errors.New("data is not pointer")
	}
	if len(data) == 0 {
		return nil
	}
	ii := sv.Elem().Interface()

	tStruct := reflect.TypeOf(ii).Elem()
	vStruct := reflect.New(tStruct)
	ckStruct := reflect.TypeOf(vStruct.Elem().Interface())
	for i := 0; i < ckStruct.NumField(); i++ {
		f := ckStruct.Field(i)
		v := vStruct.Elem().FieldByName(f.Name)
		dataValue, ok := data[f.Name]
		if !ok {
			continue
		}
		switch f.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if reflect.Int == reflect.TypeOf(dataValue).Kind() {
				v.SetInt(int64(dataValue.(int)))
			} else if reflect.Int64 == reflect.TypeOf(dataValue).Kind() {
				v.SetInt(dataValue.(int64))
			}
		case reflect.Float32, reflect.Float64:
			if reflect.Float32 == reflect.TypeOf(dataValue).Kind() {
				v.SetFloat(float64(dataValue.(float32)))
			} else if reflect.Float64 == reflect.TypeOf(dataValue).Kind() {
				v.SetFloat(dataValue.(float64))
			}
		case reflect.String:
			v.SetString(dataValue.(string))
		case reflect.Bool:
			v.SetBool(dataValue.(bool))
		case reflect.Struct:
			if dataValue != nil && f.Type.Kind() == timeKind {
				v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
				ddataVaule := dataValue.(time.Time)
				v.Set(reflect.ValueOf(&ddataVaule).Elem())
			}
		default:
			return errors.New("data type is not match")
		}
	}
	v := sv.Elem()
	v.Set(reflect.Append(v, vStruct.Elem()))
	return nil

}
