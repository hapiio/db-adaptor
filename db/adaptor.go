package db

import (
	"context"
	"database/sql"
)

type Adapter interface {
	Connect(config Config) error
	Close() error
	Insert(ctx context.Context, table string, data map[string]interface{}) error
	Update(ctx context.Context, table string, data map[string]interface{}, condition map[string]interface{}) error
	Delete(ctx context.Context, table string, condition map[string]interface{}) error
	Find(ctx context.Context, table string, condition map[string]interface{}, limit, offset int) ([]map[string]interface{}, error)
	ExecuteRaw(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRaw(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	BeginTransaction(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
}

type Config struct {
	Type             string
	ConnectionString string
	MaxOpenConns     int
	MaxIdleConns     int
}
