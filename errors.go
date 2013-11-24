package db

type Error struct {
	msg string
}

func (err *Error) Error() string {
	return err.msg
}

func (err *Error) String() string {
	return err.msg
}

var (
	ErrTableInfoError    = &Error{"cannot read table info"}
	ErrInvalidItem       = &Error{"invalid mapping item"}
	ErrNotFound          = &Error{"not found"}
	ErrMultipleRowsFound = &Error{"multiple rows found"}
)
