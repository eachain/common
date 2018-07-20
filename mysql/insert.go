package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/go-sql-driver/mysql"
)

var commonInitialismsReplacer *strings.Replacer

func init() {
	var commonInitialisms = []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "TTL", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	var commonInitialismsForReplacer []string
	for _, initialism := range commonInitialisms {
		commonInitialismsForReplacer = append(commonInitialismsForReplacer, initialism, strings.Title(strings.ToLower(initialism)))
	}
	commonInitialismsReplacer = strings.NewReplacer(commonInitialismsForReplacer...)
}

func toSnake(s string) string {
	s = commonInitialismsReplacer.Replace(s)
	rs := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsUpper(r) && len(rs) > 0 {
			rs = append(rs, '_')
		}
		rs = append(rs, unicode.ToLower(r))
	}
	return string(rs)
}

func mapToPairs(val reflect.Value) ([]string, []interface{}) {
	keys := val.MapKeys()
	if len(keys) == 0 {
		return nil, nil
	}

	fields := make([]string, len(keys))
	values := make([]interface{}, len(keys))
	for i := 0; i < len(keys); i++ {
		fields[i] = keys[i].String()
		values[i] = val.MapIndex(keys[i]).Interface()
	}
	return fields, values
}

func fieldNameEmpty(fieldName, tag string) (name string, omitempty bool) {
	name = toSnake(fieldName)
	tags := strings.Split(tag, ",")
	if len(tags) > 0 && tags[0] != "" {
		name = tags[0]
	}
	if len(tags) > 1 && tags[1] == "omitempty" {
		omitempty = true
	}
	return
}

func isEmpty(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return -1e-6 <= val.Float() && val.Float() <= 1e-6
	case reflect.String:
		return val.String() == ""
	case reflect.Struct:
		t, ok := val.Interface().(time.Time)
		if ok {
			return t.IsZero()
		}
	}
	return false
}

func structToPairs(val reflect.Value) ([]string, []interface{}) {
	typ := val.Type()
	n := val.NumField()
	fields := make([]string, 0, n)
	values := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		fv := val.Field(i)
		if !fv.CanInterface() { // unexported field
			continue
		}
		ft := typ.Field(i)
		name, omitempty := fieldNameEmpty(ft.Name, ft.Tag.Get("db"))
		if omitempty && isEmpty(fv) {
			continue
		}

		fields = append(fields, name)
		values = append(values, fv.Interface())
	}
	return fields, values
}

func insertSQL(table string, v interface{}) (string, []interface{}) {
	var fields []string
	var values []interface{}
	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Map:
		fields, values = mapToPairs(val)
	case reflect.Struct:
		fields, values = structToPairs(val)
	default:
		panic(fmt.Errorf("DB: invalid insert type: %v", val.Type().String()))
	}

	if len(fields) == 0 {
		panic(fmt.Errorf("DB: nothing to insert"))
	}

	marks := make([]string, len(values))
	for i := 0; i < len(marks); i++ {
		marks[i] = "?"
	}
	sql := "INSERT INTO `" + table + "` (" + fieldsString(fields) + ") " +
		"VALUES (" + strings.Join(marks, ", ") + ")"

	return sql, values
}

func insertManySQL(table string, v interface{}) (string, []interface{}) {
	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
	default:
		panic(fmt.Errorf("DB: invalid type: %v", val.Type().String()))
	}
	typ := val.Type().Elem()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("DB: insert many must be a slice of structure"))
	}

	length := val.Len()
	if length == 0 {
		panic(fmt.Errorf("DB: nothing to insert"))
	}

	fields, values := structToPairs(val.Index(0))
	for i := 1; i < length; i++ {
		fs, vs := structToPairs(reflect.Indirect(val.Index(i)))
		if len(fs) != len(fields) {
			panic("DB: structure fields size not equal")
		}
		values = append(values, vs...)
	}

	marks := make([]string, len(fields))
	for i := 0; i < len(marks); i++ {
		marks[i] = "?"
	}
	valMarks := "(" + strings.Join(marks, ", ") + ")"

	sql := "INSERT INTO `" + table + "` (" + fieldsString(fields) + ") " +
		"VALUES " + valMarks
	for i := 1; i < length; i++ {
		sql += ", " + valMarks
	}

	return sql, values
}

func fieldsString(fields []string) string {
	return "`" + strings.Join(fields, "`, `") + "`"
}

/*
InsertRow v只允许map[string]interface{}和struct类型,
类型内部不允许嵌套复杂类型。用法:

	t := &T{...}
	result, err := mysql.InsertRow("biz_name", "table_name", t)
	if err != nil {
		...
	}
	_ = result
*/
func InsertRow(biz, table string, v interface{}) (sql.Result, error) {
	db, err := Get(biz)
	if err != nil {
		return nil, err
	}
	sql, values := insertSQL(table, v)
	return db.Exec(sql, values...)
}

/*
InsertRows 参数v必须是slice of struct类型。
在a slice of struct中，不要某些用omitempty默认值，有些不用。
否则有可能会因为字段对不上而panic。用法:

	type T struct {
		ID int64 `db:"id,omitempty"` // 除了id，其他字段最好不用omitempty
		...
	}

	ts := []T{...} // or []*T{...}
	result, err := InsertRows("biz_name", "table_name", ts)
	if err != nil {
		...
	}
	_ = result
*/
func InsertRows(biz, table string, v interface{}) (sql.Result, error) {
	db, err := Get(biz)
	if err != nil {
		return nil, err
	}
	sql, values := insertManySQL(table, v)
	return db.Exec(sql, values...)
}

/*
IsDup 用于判断mysql insert返回的error是不是duplicate,
即出现primary/unique冲突。
例:

	_, err := InsertRow(biz, table, &T{...})
	if err != nil {
		if IsDup(err) {
			...
		}
	}
*/
func IsDup(err error) bool {
	if err == nil {
		return false
	}

	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}

	const (
		ER_DUP_ENTRY  = 1062
		ER_DUP_UNIQUE = 1169
	)
	return me.Number == ER_DUP_ENTRY || me.Number == ER_DUP_UNIQUE
}
