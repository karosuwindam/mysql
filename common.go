package mysql

import (
	"reflect"
	"time"
)

const (
	// TimeLayout String変換用テンプレート
	TimeLayout = "2006-01-02 15:04:05.999999999"
	// TimeLayout2 String変換用テンプレート
	TimeLayout2 = "2006-01-02 15:04:05.99999999 +0000 UTC"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()
