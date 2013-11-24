package db

import (
	"reflect"
	"strings"
)

type TableMapping struct {
	name    string
	session *Session
}

func (m *TableMapping) Query() *Query {
	return &Query{
		mapping:    m,
		filtercond: make([]string, 0, 2),
		filtervals: make([]interface{}, 0, 2),
		limit:      -1,
		offset:     -1,
		order:      order{asc: make([]string, 0, 1), desc: make([]string, 0, 1)},
	}
}

func (m *TableMapping) Save(item interface{}) (created bool, err error) {
	table, err := m.tableinfo()
	if err != nil {
		return false, err
	}
	val := reflect.ValueOf(item)
	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Type().Kind() != reflect.Struct {
		return false, ErrInvalidItem
	}
	var pkfield reflect.Value
	fieldNames := make([]string, 0, len(table.fields))
	args := make([]interface{}, 0, len(table.fields))

	for _, field := range table.fields {
		f := val.FieldByName(field.name)
		if f.IsValid() {
			if field.pk {
				pkfield = f
				continue
			}
			fieldNames = append(fieldNames, field.name)
			args = append(args, f.Addr().Interface())
		}
	}

	// XXX what if struct does not map table completly?
	if pkfield.IsValid() && pkfield.Interface() != reflect.Zero(pkfield.Type()).Interface() {
		created = false
		sqlChunks := []string{`UPDATE "`, table.name, `" SET "`}
		for i, name := range fieldNames {
			if i == len(fieldNames)-1 {
				sqlChunks = append(sqlChunks, name, `" = ?`)
			} else {
				sqlChunks = append(sqlChunks, name, `" = ?, "`)
			}
		}
		sqlChunks = append(sqlChunks, ` WHERE "`, table.pkfield.name, `" = ?`)
		sql := strings.Join(sqlChunks, "")
		// value for WHERE <primary key> = ?
		args = append(args, pkfield.Addr().Interface())
		_, err := m.session.Exec(sql, args...)
		if err != nil {
			return created, err
		}
	} else {
		created = true
		sqlChunks := []string{`INSERT INTO `, table.name, `("`}
		for i, name := range fieldNames {
			if i == len(fieldNames)-1 {
				sqlChunks = append(sqlChunks, name, `") VALUES(`)
			} else {
				sqlChunks = append(sqlChunks, name, `", `)
			}
		}
		for i := len(fieldNames); i > 0; i-- {
			if i == 1 {
				sqlChunks = append(sqlChunks, "?)")
			} else {
				sqlChunks = append(sqlChunks, "?, ")
			}
		}
		sql := strings.Join(sqlChunks, "")
		res, err := m.session.Exec(sql, args...)
		if err != nil {
			return created, err
		}
		id, err := res.LastInsertId()
		// we cannot assume that the primary key is number, so just skip this one
		if err == nil {
			pkfield.SetInt(id)
		}
	}

	return created, nil
}

func (m *TableMapping) Delete(item interface{}) error {
	table, err := m.tableinfo()
	if err != nil {
		return err
	}

	val := reflect.ValueOf(item)
	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Type().Kind() != reflect.Struct {
		return ErrInvalidItem
	}
	pkval := val.FieldByName(table.pkfield.name)
	if !pkval.IsValid() {
		return ErrInvalidItem
	}

	sqlChunks := []string{`DELETE FROM "`, table.name, `" WHERE "`, table.pkfield.name, `" = ?`}
	sql := strings.Join(sqlChunks, "")
	res, err := m.session.Exec(sql, pkval.Interface())
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNotFound
	}
	return nil
}

func (m *TableMapping) tableinfo() (*tableinfo, error) {
	table, err := m.session.dialect.TableInfo(m.session.db, m.name)
	if err != nil {
		m.session.log.Error("cannot acquire table info: %s", err)
		return nil, ErrTableInfoError
	}
	return table, nil
}
