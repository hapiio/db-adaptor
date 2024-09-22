package db

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSQLAdapter(t *testing.T) {
	// Test both MySQL and PostgreSQL
	databases := []struct {
		name             string
		dbType           string
		connectionString string
	}{
		{
			name:             "MySQL",
			dbType:           "mysql",
			connectionString: os.Getenv("MYSQL_URI"),
		},
		{
			name:             "PostgreSQL",
			dbType:           "postgres",
			connectionString: os.Getenv("POSTGRES_URI"),
		},
	}

	for _, db := range databases {
		t.Run(db.name, func(t *testing.T) {
			if db.connectionString == "" {
				switch db.dbType {
				case "mysql":
					db.connectionString = "testuser:testpassword@tcp(localhost:3306)/testdb"
				case "postgres":
					db.connectionString = "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable"
				}
			}

			config := Config{
				Type:             db.dbType,
				ConnectionString: db.connectionString,
				MaxOpenConns:     10,
				MaxIdleConns:     5,
			}

			adapter, err := NewAdapter(config)
			if err != nil {
				t.Fatalf("Failed to create %s adapter: %v", db.name, err)
			}
			defer adapter.Close()

			// Wait for database to be ready
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err = waitForSQL(ctx, adapter)
			if err != nil {
				t.Fatalf("%s is not ready: %v", db.name, err)
			}

			// Create test table
			_, err = adapter.ExecuteRaw(ctx, `
				CREATE TABLE IF NOT EXISTS test_table (
					id SERIAL PRIMARY KEY,
					name VARCHAR(100),
					age INT
				)
			`)
			if err != nil {
				t.Fatalf("Failed to create test table: %v", err)
			}

			t.Run("TestInsert", func(t *testing.T) {
				err := adapter.Insert(ctx, "test_table", map[string]interface{}{
					"name": "John Doe",
					"age":  30,
				})
				if err != nil {
					t.Fatalf("Insert failed: %v", err)
				}
			})

			t.Run("TestFind", func(t *testing.T) {
				results, err := adapter.Find(ctx, "test_table", map[string]interface{}{"name": "John Doe"}, 10, 0)
				if err != nil {
					t.Fatalf("Find failed: %v", err)
				}
				if len(results) != 1 {
					t.Fatalf("Expected 1 result, got %d", len(results))
				}
			})

			t.Run("TestUpdate", func(t *testing.T) {
				err := adapter.Update(ctx, "test_table", map[string]interface{}{"age": 31}, map[string]interface{}{"name": "John Doe"})
				if err != nil {
					t.Fatalf("Update failed: %v", err)
				}
			})

			t.Run("TestDelete", func(t *testing.T) {
				err := adapter.Delete(ctx, "test_table", map[string]interface{}{"name": "John Doe"})
				if err != nil {
					t.Fatalf("Delete failed: %v", err)
				}
			})

			t.Run("TestTransaction", func(t *testing.T) {
				tx, err := adapter.BeginTransaction(ctx)
				if err != nil {
					t.Fatalf("BeginTransaction failed: %v", err)
				}

				err = adapter.Insert(ctx, "test_table", map[string]interface{}{
					"name": "Jane Doe",
					"age":  25,
				})
				if err != nil {
					t.Fatalf("Insert in transaction failed: %v", err)
				}

				err = tx.Commit()
				if err != nil {
					t.Fatalf("Transaction commit failed: %v", err)
				}

				results, err := adapter.Find(ctx, "test_table", map[string]interface{}{"name": "Jane Doe"}, 10, 0)
				if err != nil {
					t.Fatalf("Find after transaction failed: %v", err)
				}
				if len(results) != 1 {
					t.Fatalf("Expected 1 result after transaction, got %d", len(results))
				}
			})

			// Clean up
			_, err = adapter.ExecuteRaw(ctx, "DROP TABLE IF EXISTS test_table")
			if err != nil {
				t.Fatalf("Failed to drop test table: %v", err)
			}
		})
	}
}

func waitForSQL(ctx context.Context, adapter Adapter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := adapter.ExecuteRaw(ctx, "SELECT 1")
			if err == nil {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
