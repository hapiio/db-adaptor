package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" // for MySQL
	_ "github.com/lib/pq"              // for PostgreSQL
)

type SQLAdapter struct {
	db *sql.DB
}

func (s *SQLAdapter) Connect(connectionString string, driverName string) error {
	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *SQLAdapter) Insert(query string, args ...interface{}) error {
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *SQLAdapter) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

func (s *SQLAdapter) Update(ctx context.Context, table string, set map[string]interface{}, where map[string]interface{}) error {
	setValues := make([]string, 0)
	whereValues := make([]string, 0)
	args := make([]interface{}, 0)

	for key, value := range set {
		setValues = append(setValues, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	for key, value := range where {
		whereValues = append(whereValues, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setValues, ", "), strings.Join(whereValues, " AND "))
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *SQLAdapter) Delete(ctx context.Context, table string, where map[string]interface{}) error {
	whereValues := make([]string, 0)
	args := make([]interface{}, 0)

	for key, value := range where {
		whereValues = append(whereValues, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, strings.Join(whereValues, " AND "))
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *SQLAdapter) Find(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *SQLAdapter) BatchInsert(ctx context.Context, table string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}

	keys := make([]string, 0)
	for key := range records[0] {
		keys = append(keys, key)
	}

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	for _, record := range records {
		valueStrings = append(valueStrings, "("+strings.Repeat("?,", len(keys)-1)+"?)")
		for _, key := range keys {
			valueArgs = append(valueArgs, record[key])
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, strings.Join(keys, ", "), strings.Join(valueStrings, ", "))
	_, err := s.db.ExecContext(ctx, query, valueArgs...)
	return err
}

func (s *SQLAdapter) BatchUpdate(ctx context.Context, table string, records []map[string]interface{}, primaryKey string) error {
	// Assuming all records have the same fields and primary key is present
	if len(records) == 0 {
		return nil
	}

	keys := make([]string, 0)
	for key := range records[0] {
		if key != primaryKey {
			keys = append(keys, key)
		}
	}

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	for _, record := range records {
		valueStrings = append(valueStrings, "("+strings.Repeat("?,", len(records[0])-1)+"?)")
		for _, key := range keys {
			valueArgs = append(valueArgs, record[key])
		}
		valueArgs = append(valueArgs, record[primaryKey])
	}

	setStatements := make([]string, 0)
	for _, key := range keys {
		setStatements = append(setStatements, fmt.Sprintf("%s = VALUES(%s)", key, key))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES %s ON DUPLICATE KEY UPDATE %s",
		table, strings.Join(keys, ", "), primaryKey,
		strings.Join(valueStrings, ", "), strings.Join(setStatements, ", "))
	_, err := s.db.ExecContext(ctx, query, valueArgs...)
	return err
}

func (s *SQLAdapter) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

func (s *SQLAdapter) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
