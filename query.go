package db

import (
	"fmt"
	"reflect"
	"strings"
)

type Query struct {
	mapping    *TableMapping
	filtercond []string
	filtervals []interface{}
	limit      int64
	offset     int64
	order      order
}

type order struct {
	asc  []string
	desc []string
}

func (q *Query) Where(cond string, val interface{}) *Query {
	q.filtercond = append(q.filtercond, cond)
	q.filtervals = append(q.filtervals, val)
	return q
}

func (q *Query) OrderBy(fields ...string) *Query {
	q.order.asc = append(q.order.asc, fields...)
	return q
}

func (q *Query) OrderDesc(fields ...string) *Query {
	q.order.desc = append(q.order.desc, fields...)
	return q
}

func (q *Query) Limit(limit int64) *Query {
	q.limit = limit
	return q
}

func (q *Query) Offset(offset int64) *Query {
	q.offset = offset
	return q
}

func (q *Query) One(dest interface{}) error {
	table, err := q.mapping.tableinfo()
	if err != nil {
		return err
	}
	structval := reflect.ValueOf(dest)
	for structval.Type().Kind() == reflect.Ptr {
		structval = structval.Elem()
	}
	if structval.Type().Kind() != reflect.Struct {
		return ErrInvalidItem
	}
	sqlargs := q.sqlargs(table, structval)
	if len(sqlargs) == 0 {
		return ErrInvalidItem
	}
	sqlquery := q.sqlquery(table, structval)
	rows, err := q.mapping.session.Query(sqlquery, q.filtervals...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrNotFound
	}
	if err := rows.Scan(sqlargs...); err != nil {
		return err
	}
	if rows.Next() {
		return ErrMultipleRowsFound
	}

	return nil
}

func (q *Query) All(dest interface{}) error {
	table, err := q.mapping.tableinfo()
	if err != nil {
		return err
	}
	slice := reflect.ValueOf(dest)
	for slice.Type().Kind() == reflect.Ptr {
		slice = slice.Elem()
	}
	if slice.Type().Kind() != reflect.Slice {
		q.mapping.session.log.Error(
			"query All() destination is not slice (got %s)", slice.Type().Kind())
		return ErrInvalidItem
	}
	itemTpIsPtr := false
	itemTp := slice.Type().Elem()
	if itemTp.Kind() == reflect.Ptr {
		itemTpIsPtr = true
		itemTp = itemTp.Elem()
	}
	if itemTp.Kind() != reflect.Struct {
		q.mapping.session.log.Error(
			"query All() destination is not slice of structures (slice of %s)", itemTp.Kind())
		return ErrInvalidItem
	}

	sqlquery := q.sqlquery(table, reflect.New(itemTp).Elem())
	rows, err := q.mapping.session.Query(sqlquery, q.filtervals...)
	if err != nil {
		q.mapping.session.log.Error("select query error: %s\n%s", err, sqlquery)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		structval := reflect.New(itemTp)
		sqlargs := q.sqlargs(table, structval.Elem())
		if len(sqlargs) == 0 {
			q.mapping.session.log.Error("query destination does not map source table")
			return ErrInvalidItem
		}
		if err := rows.Scan(sqlargs...); err != nil {
			q.mapping.session.log.Error("cannot scan result row: %s", err)
			return err
		}
		if itemTpIsPtr {
			slice.Set(reflect.Append(slice, structval))
		} else {
			slice.Set(reflect.Append(slice, reflect.Indirect(structval)))
		}
	}
	return nil
}

func (q *Query) Count() (int64, error) {
	return 0, nil
}

func (q *Query) Exists() (bool, error) {
	return false, nil
}

func (q *Query) sqlargs(table *tableinfo, structval reflect.Value) (args []interface{}) {
	args = make([]interface{}, 0, structval.NumField())
	for _, field := range table.fields {
		f := structval.FieldByName(field.name)
		if f.IsValid() {
			args = append(args, f.Addr().Interface())
		}
	}
	return args
}

func (q *Query) sqlquery(table *tableinfo, structval reflect.Value) (sql string) {
	sqlChunks := make([]string, 0, 12)

	sqlChunks = append(sqlChunks, `SELECT "`)
	for _, field := range table.fields {
		f := structval.FieldByName(field.name)
		if f.IsValid() {
			sqlChunks = append(sqlChunks, field.dbname, `", "`)
		}
	}
	sqlChunks[len(sqlChunks)-1] = `" FROM `
	sqlChunks = append(sqlChunks, `"`, table.name, `"`)

	if len(q.filtercond) != 0 {
		sqlChunks = append(sqlChunks, ` WHERE `)
		for i, cond := range q.filtercond {
			sqlChunks = append(sqlChunks, cond, `?`)
			if i != len(q.filtercond)-1 {
				sqlChunks = append(sqlChunks, cond, ` AND `)
			}
		}
	}

	if len(q.order.asc) != 0 || len(q.order.desc) != 0 {
		sqlChunks = append(sqlChunks, " ORDER BY ")
		for _, name := range q.order.asc {
			sqlChunks = append(sqlChunks, `"`, name, `" ASC`, `, `)
		}
		for _, name := range q.order.desc {
			sqlChunks = append(sqlChunks, `"`, name, `" DESC`, `, `)
		}
		sqlChunks[len(sqlChunks) - 1] = " "
	}

	if q.limit > -1 {
		sqlChunks = append(sqlChunks, fmt.Sprintf(` LIMIT %d `, q.limit))
	}
	if q.offset > -1 {
		sqlChunks = append(sqlChunks, fmt.Sprintf(` OFFSET %d `, q.offset))
	}

	return strings.Join(sqlChunks, "")
}
