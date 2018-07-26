package mysql

import (
	gosql "database/sql"
	"strings"
)

type UpdateSQL struct {
	table     string
	sets      []string
	setMarks  []interface{}
	conds     []string
	condMarks []interface{}
}

func (sql *UpdateSQL) clone() *UpdateSQL {
	new := *sql
	return &new
}

func (sql *UpdateSQL) Update(table string) *UpdateSQL {
	sql = sql.clone()
	sql.table = table
	return sql
}

func (sql *UpdateSQL) Set(field string, val interface{}) *UpdateSQL {
	sql = sql.clone()
	sql.sets = append(sql.sets, field)
	sql.setMarks = append(sql.setMarks, val)
	return sql
}

func (sql *UpdateSQL) Where(cond string, marks ...interface{}) *UpdateSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, cond)
	sql.condMarks = append(sql.condMarks, marks...)
	return sql
}

func (sql *UpdateSQL) And(cond string, marks ...interface{}) *UpdateSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, "AND", cond)
	sql.condMarks = append(sql.condMarks, marks...)
	return sql
}

func (sql *UpdateSQL) Or(cond string, marks ...interface{}) *UpdateSQL {
	sql = sql.clone()
	sql.conds = append(sql.conds, "OR", cond)
	sql.condMarks = append(sql.condMarks, marks...)
	return sql
}

func (sql *UpdateSQL) String() string {
	query := "UPDATE `" + sql.table + "` SET "
	sets := make([]string, 0, len(sql.sets))
	for _, set := range sql.sets {
		if strings.Contains(set, "=") &&
			strings.Contains(set, "?") {
			sets = append(sets, set)
		} else {
			sets = append(sets, set+" = ?")
		}
	}
	query += strings.Join(sets, ", ")
	if len(sql.conds) > 0 {
		query += " WHERE "
		query += strings.Join(sql.conds, " ")
	}
	return query
}

func (sql *UpdateSQL) Marks() []interface{} {
	return append(sql.setMarks, sql.condMarks...)
}

func (sql *UpdateSQL) Exec(biz string) (gosql.Result, error) {
	db, err := Get(biz)
	if err != nil {
		return nil, err
	}
	return db.Exec(sql.String(), sql.Marks()...)
}
