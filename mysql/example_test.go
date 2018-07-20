package mysql_test

import (
	"database/sql"
	"log"
	"time"

	"mysql"
)

var db *sql.DB

type T struct {
	A int
	B string
	C time.Time
}

func ExampleAllFields() {
	var t T // test structure
	names, fields := AllFields(&t)
	sql := "SELECT " + names + " FROM table LIMIT 1"
	err := db.QueryRow(sql).Scan(fields...)
	if err != nil {
		log.Fatal(err)
	}
	_ = t // all fields are filled
}

func ExampleSelectSQL() {
	var err error
	query := new(SelectSQL).From(table).Where("id > ?", 123).OrderBy("id desc")

	var t T
	err = query.Limit(0, 1).Row(biz, &t)
	if err != nil {
		log.Fatal(err)
	}

	var ts []T // or []*T
	err = query.Rows(biz, &ts)
	if err != nil {
		log.Fatal(err)
	}

	var maxId, count int64
	err = query.Select("max(id), count(1)").Row2(biz, &maxId, &count)
	if err != nil {
		log.Fatal(err)
	}

	// 如果只选取其中几个字段, 请用结构体
	var names []struct{ Name string }
	err = query.Select("name").Rows(biz, &names)
	if err != nil {
		log.Fatal(err)
	}
}
