package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type SQLAdapter struct {
	db     *sql.DB
	dbType string
}

func (s *SQLAdapter) Connect(config Config) error {
	db, err := sql.Open(config.Type, config.ConnectionString)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	s.db = db
	s.dbType = config.Type
	return nil
}

func (s *SQLAdapter) Insert(ctx context.Context, table string, data map[string]interface{}) error {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for column, value := range data {
		columns = append(columns, column)
		values = append(values, value)
		if s.dbType == "postgres" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		} else {
			placeholders = append(placeholders, "?")
		}
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	_, err := s.db.ExecContext(ctx, query, values...)
	return err
}

func (s *SQLAdapter) Find(ctx context.Context, table string, filter map[string]interface{}, limit, offset int) ([]map[string]interface{}, error) {
	whereClause, values := s.buildWhereClause(filter)
	query := fmt.Sprintf("SELECT * FROM %s", table)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	if limit > 0 {
		if s.dbType == "postgres" {
			query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
		} else {
			query += fmt.Sprintf(" LIMIT %d, %d", offset, limit)
		}
	}

	rows, err := s.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanRows(rows)
}

func (s *SQLAdapter) Update(ctx context.Context, table string, set map[string]interface{}, where map[string]interface{}) error {
	setClause, setValues := s.buildSetClause(set)
	whereClause, whereValues := s.buildWhereClause(where)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setClause, whereClause)
	values := append(setValues, whereValues...)
	_, err := s.db.ExecContext(ctx, query, values...)
	return err
}

func (s *SQLAdapter) Delete(ctx context.Context, table string, where map[string]interface{}) error {
	whereClause, values := s.buildWhereClause(where)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereClause)
	_, err := s.db.ExecContext(ctx, query, values...)
	return err
}

func (s *SQLAdapter) ExecuteRaw(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

func (s *SQLAdapter) QueryRaw(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *SQLAdapter) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &SQLTransaction{tx: tx}, nil
}

func (s *SQLAdapter) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLAdapter) buildSetClause(set map[string]interface{}) (string, []interface{}) {
	var clauses []string
	var values []interface{}
	i := 1
	for column, value := range set {
		if s.dbType == "postgres" {
			clauses = append(clauses, fmt.Sprintf("%s = $%d::%s", column, i, s.getPostgresType(value)))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s = ?", column))
		}
		values = append(values, value)
		i++
	}
	return strings.Join(clauses, ", "), values
}

func (s *SQLAdapter) buildWhereClause(filter map[string]interface{}) (string, []interface{}) {
	if len(filter) == 0 {
		return "", nil
	}

	var clauses []string
	var values []interface{}
	i := 1
	for column, value := range filter {
		if s.dbType == "postgres" {
			clauses = append(clauses, fmt.Sprintf("%s = $%d::%s", column, i, s.getPostgresType(value)))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s = ?", column))
		}
		values = append(values, value)
		i++
	}
	return strings.Join(clauses, " AND "), values
}

func (s *SQLAdapter) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, column := range columns {
			row[column] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

type SQLTransaction struct {
	tx *sql.Tx
}

func (t *SQLTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *SQLTransaction) Rollback() error {
	return t.tx.Rollback()
}

func (s *SQLAdapter) getPostgresType(value interface{}) string {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "numeric"
	case reflect.Bool:
		return "boolean"
	default:
		return "text"
	}
}
