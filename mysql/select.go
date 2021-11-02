package mysql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/*
SelectSQL用于select语句, 函数名和select语法保持一致,
参数和database/sql一样, 用?标记.
SelectSQL可以链式调用, 每次返回一个浅拷贝,
所以可以被复用.
其他的UpdateSQL、DeleteSQL同理. 用法:

	var ts []*T
	err := new(mysql.SelectSQL).From(table).
		Where("id = ?", id).
		And("name = ?", name).
		OrderBy("id desc").
		Limit(100, 10).
		Rows(biz, &ts)
	if err != nil {
		...
	}
*/
type SelectSQL struct {
	fields string
	table  string
	conds  []string
	marks  []interface{}
	order  string
	offset int
	limit  int
}

func (sql *SelectSQL) clone() *SelectSQL {
	new := *sql
	return &new
}

func (sql *SelectSQL) Select(fields string) *SelectSQL {
	sql = sql.clone()
	sql.fields = fields
	return sql
}

func (sql *SelectSQL) From(table string) *SelectSQL {
	sql = sql.clone()
	sql.table = table
	return sql
}

func (sql *SelectSQL) Where(cond string, marks ...interface{}) *SelectSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, cond)
	sql.marks = append(sql.marks, marks...)
	return sql
}

func (sql *SelectSQL) And(cond string, marks ...interface{}) *SelectSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, "AND", cond)
	sql.marks = append(sql.marks, marks...)
	return sql
}

func (sql *SelectSQL) Or(cond string, marks ...interface{}) *SelectSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, "OR", cond)
	sql.marks = append(sql.marks, marks...)
	return sql
}

func (sql *SelectSQL) OrderBy(order string) *SelectSQL {
	sql = sql.clone()
	sql.order = order
	return sql
}

func (sql *SelectSQL) Limit(offset, limit int) *SelectSQL {
	sql = sql.clone()
	sql.offset = offset
	sql.limit = limit
	return sql
}

/*
String 返回sql语句，需要配合Marks一起用。用法：

	var db *gosql.DB // import gosql "database/sql"
	sql := new(mysql.SelectSQL).XXX.XXX...
	row := db.QueryRow(sql.String(), sql.Marks()...)
	_ = row
*/
func (sql *SelectSQL) String() string {
	// query := "SELECT " + sql.fields + " FROM `" + sql.table + "`"
	query := "SELECT " + sql.fields + " FROM " + sql.table
	if len(sql.conds) > 0 {
		query += " WHERE "
		query += strings.Join(sql.conds, " ")
	}
	if sql.order != "" {
		query += " ORDER BY " + sql.order
	}
	if sql.limit != 0 {
		query += " LIMIT "
		if sql.offset != 0 {
			query += strconv.FormatInt(int64(sql.offset), 10) + ", "
		}
		query += strconv.FormatInt(int64(sql.limit), 10)
	}
	return query
}

// Marks 返回所有被标记为?的值，参考String.
func (sql *SelectSQL) Marks() []interface{} {
	return sql.marks
}

// Row 用于只返回一条记录的sql语句。dst应该传structure.
func (sql *SelectSQL) Row(biz string, dst interface{}) error {
	names, fields := AllFields(dst)
	if fields := strings.TrimSpace(sql.fields); fields == "" || fields == "*" {
		sql = sql.Select(names)
	}
	return queryRow(biz, sql.String(), sql.marks, fields)
}

/*
Row2 用于只返回一条记录，但列是自定义的。例如：

	var count, maxId int64
	err := new(mysql.SelectSQL).Select("count(1), max(id)").From(table).Row2(biz, &count, &maxId)
	if err != nil {
		...
	}
*/
func (sql *SelectSQL) Row2(biz string, fields ...interface{}) error {
	return queryRow(biz, sql.String(), sql.marks, fields)
}

/*
Rows 用于返回多条记录。将记录写入dst。dst应该是个 slice of structure。用法：

	var ts []T // or []*T
	err := new(mysql.SelectSQL).From(table).Limit(0, 5).Rows(biz, &ts)
	if err != nil {
		...
	}
*/
func (sql *SelectSQL) Rows(biz string, dst interface{}) error {
	if fields := strings.TrimSpace(sql.fields); fields == "" || fields == "*" {
		sql = sql.Select(selectFieldNames(dst))
	}
	return queryRows(biz, sql.String(), sql.marks, dst)
}

/*
AllFields 用于select语句，返回fieldNames和fieldValues。
fieldNames用于生成sql，
fieldValues用于row.Scan(values...)。
fieldNames和fieldValues是一一对应的。
用法:

	var t T
	names, fields := AllFields(&t)
	sql := "SELECT " + names + " FROM " + table + " LIMIT 1"
	err := db.QueryRow(sql).Scan(fields...)
*/
func AllFields(v interface{}) (string, []interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		panic(fmt.Errorf("DB: select all need a pointer"))
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		panic(fmt.Errorf("DB: select all can only be a structure"))
	}

	typ := val.Type()
	n := val.NumField()
	fields := make([]string, 0, n)
	values := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		fv := val.Field(i)
		if !fv.CanInterface() || !fv.CanAddr() || !fv.CanSet() {
			continue
		}
		ft := typ.Field(i)
		name, _ := fieldNameEmpty(ft.Name, ft.Tag.Get("db"))

		fields = append(fields, name)
		values = append(values, fv.Addr().Interface())
	}
	return fieldsString(fields), values
}

func queryRow(biz, sql string, marks, fields []interface{}) error {
	db, err := Get(biz)
	if err != nil {
		return err
	}

	return db.QueryRow(sql, marks...).Scan(fields...)
}

/*
queryRows dst must be a slice, like: `&[]T` or `&[]*T`.
It should be called like this:

	var dst []T // or []*T
	err := queryRows(biz, sql, marks, &dst)
	if err != nil {
		...
	}
*/
func queryRows(biz, sql string, marks []interface{}, dst interface{}) error {
	db, err := Get(biz)
	if err != nil {
		return err
	}

	rows, err := db.Query(sql, marks...)
	if err != nil {
		return err
	}
	defer rows.Close()

	origin := reflect.ValueOf(dst).Elem()
	slice := origin
	elemTyp := slice.Type().Elem()
	isPtr := false
	if elemTyp.Kind() == reflect.Ptr {
		elemTyp = elemTyp.Elem()
		isPtr = true
	}

	for rows.Next() {
		elem := reflect.New(elemTyp)
		_, fields := AllFields(elem.Interface())
		err = rows.Scan(fields...)
		if err != nil {
			return err
		}
		if isPtr {
			slice = reflect.Append(slice, elem)
		} else {
			slice = reflect.Append(slice, elem.Elem())
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	origin.Set(slice)
	return nil
}

func selectFieldNames(slice interface{}) string {
	elemTyp := reflect.ValueOf(slice).Elem().Type().Elem()
	if elemTyp.Kind() == reflect.Ptr {
		elemTyp = elemTyp.Elem()
	}
	elem := reflect.New(elemTyp)
	names, _ := AllFields(elem.Interface())
	return names
}
