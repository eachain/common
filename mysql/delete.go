package mysql

import (
	gosql "database/sql"
	"strconv"
	"strings"
)

type DeleteSQL struct {
	table string
	conds []string
	marks []interface{}
	order string
	limit int
}

func (sql *DeleteSQL) clone() *DeleteSQL {
	new := *sql
	return &new
}

func (sql *DeleteSQL) From(table string) *DeleteSQL {
	sql = sql.clone()
	sql.table = table
	return sql
}

func (sql *DeleteSQL) Where(cond string, marks ...interface{}) *DeleteSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, cond)
	sql.marks = append(sql.marks, marks...)
	return sql
}

func (sql *DeleteSQL) And(cond string, marks ...interface{}) *DeleteSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, "AND", cond)
	sql.marks = append(sql.marks, marks...)
	return sql
}

func (sql *DeleteSQL) Or(cond string, marks ...interface{}) *DeleteSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, "OR", cond)
	sql.marks = append(sql.marks, marks...)
	return sql
}

func (sql *DeleteSQL) OrderBy(order string) *DeleteSQL {
	sql = sql.clone()
	sql.order = order
	return sql
}

func (sql *DeleteSQL) Limit(limit int) *DeleteSQL {
	sql = sql.clone()
	sql.limit = limit
	return sql
}

func (sql *DeleteSQL) String() string {
	query := "DELETE FROM `" + sql.table + "`"
	if len(sql.conds) > 0 {
		query += " WHERE "
		query += strings.Join(sql.conds, " ")
	}
	if sql.order != "" {
		query += " ORDER BY " + sql.order
	}
	if sql.limit != 0 {
		query += " LIMIT " + strconv.FormatInt(int64(sql.limit), 10)
	}
	return query
}

func (sql *DeleteSQL) Marks() []interface{} {
	return sql.marks
}

func (sql *DeleteSQL) Exec(biz string) (gosql.Result, error) {
	db, err := Get(biz)
	if err != nil {
		return nil, err
	}
	return db.Exec(sql.String(), sql.Marks()...)
}
