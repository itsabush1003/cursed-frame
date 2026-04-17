package repository

import "github.com/jmoiron/sqlx"

type WriteMethod string

const (
	Insert WriteMethod = "INSERT"
	Update WriteMethod = "UPDATE"
	Delete WriteMethod = "DELETE"
)

type WriteRequest struct {
	Table    string
	Method   WriteMethod
	Targets  []string
	Params   any
	Conds    string
	ResultCh chan<- error
}

type IDatabase interface {
	Command(string, WriteRequest)
	Query(string, string, ...any) (*sqlx.Rows, error)
	QueryRow(string, string, ...any) *sqlx.Row
	QueryIn(string, string, ...any) (*sqlx.Rows, error)
}
