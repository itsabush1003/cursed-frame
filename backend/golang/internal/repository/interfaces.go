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
	Command(dbName string, request WriteRequest)
	Query(dbName string, sql string, params ...any) (*sqlx.Rows, error)
	QueryRow(dbName string, sql string, params ...any) *sqlx.Row
	QueryIn(dbName string, sql string, params ...any) (*sqlx.Rows, error)
}
