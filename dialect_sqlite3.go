package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var (
	rxDash  = regexp.MustCompile(`\_.`)
	rxCamel = regexp.MustCompile(`[A-Z]`)
)

var Sqlite3Dialect = &sqlite3Dialect{
	tables: make(map[string]*tableinfo),
}

type sqlite3Dialect struct {
	lock   sync.RWMutex
	tables map[string]*tableinfo
}

func (d *sqlite3Dialect) TableInfo(db *sql.DB, name string) (*tableinfo, error) {
	d.lock.RLock()
	table, ok := d.tables[name]
	d.lock.RUnlock()
	if ok {
		return table, nil
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	sql := fmt.Sprintf("pragma table_info(%s)", name)
	res, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	table = &tableinfo{
		name:   name,
		fields: make([]*tablefield, 0),
	}

	// we don't care about those
	var cid, notNull, tp, defaultValue interface{}
	for res.Next() {
		f := &tablefield{}
		// cid, name, type, notnull, dflt_value, pk
		err := res.Scan(&cid, &f.dbname, &tp, &notNull, &defaultValue, &f.pk)
		if err != nil {
			return nil, err
		}
		f.name = dashToCamel(f.dbname)
		table.fields = append(table.fields, f)
		if f.pk {
			table.pkfield = f
		}
	}
	return table, nil
}

func dashToCamel(s string) string {
	camel := rxDash.ReplaceAllStringFunc(s, func(m string) string {
		return strings.ToUpper(m[1:])
	})
	camel = strings.ToUpper(camel[:1]) + camel[1:]
	return camel
}

func camelToDash(s string) string {
	dash := rxCamel.ReplaceAllStringFunc(s, func(m string) string {
		return "_" + strings.ToLower(m)
	})
	for dash[0] == '_' {
		dash = dash[1:]
	}
	return dash
}
