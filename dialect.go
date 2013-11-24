package db

import (
	"database/sql"
)

type Dialect interface {
	TableInfo(*sql.DB, string) (*tableinfo, error)
}

type tableinfo struct {
	name    string
	fields  []*tablefield
	pkfield *tablefield
}

type tablefield struct {
	name   string
	dbname string
	pk     bool
}
