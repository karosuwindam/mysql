package mysql

import (
	"errors"
	"reflect"
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
func (cfg *MySqlConfig) Read(tName string, data interface{}, limit map[string]string) error {
	//ポインター型でない場合はエラーを返す
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		return errors.New("data is not pointer")
	}
	return errors.New("data is not struct")
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
