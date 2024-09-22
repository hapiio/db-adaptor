package db

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMongoDBAdapter(t *testing.T) {
	// Use environment variable or default to the Docker Compose service
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://testuser:testpassword@localhost:27017"
	}

	config := Config{
		Type:             "mongodb",
		ConnectionString: mongoURI,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create MongoDB adapter: %v", err)
	}
	defer adapter.Close()

	// Wait for MongoDB to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = waitForMongoDB(ctx, adapter)
	if err != nil {
		t.Fatalf("MongoDB is not ready: %v", err)
	}

	t.Run("TestInsert", func(t *testing.T) {
		err := adapter.Insert(ctx, "test_collection", map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	})

	t.Run("TestFind", func(t *testing.T) {
		results, err := adapter.Find(ctx, "test_collection", map[string]interface{}{"name": "John Doe"}, 10, 0)
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("TestUpdate", func(t *testing.T) {
		err := adapter.Update(ctx, "test_collection", map[string]interface{}{"age": 31}, map[string]interface{}{"name": "John Doe"})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	})

	t.Run("TestDelete", func(t *testing.T) {
		err := adapter.Delete(ctx, "test_collection", map[string]interface{}{"name": "John Doe"})
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("TestTransaction", func(t *testing.T) {
		tx, err := adapter.BeginTransaction(ctx)
		if err != nil {
			t.Fatalf("BeginTransaction failed: %v", err)
		}

		err = adapter.Insert(ctx, "test_collection", map[string]interface{}{
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

		results, err := adapter.Find(ctx, "test_collection", map[string]interface{}{"name": "Jane Doe"}, 10, 0)
		if err != nil {
			t.Fatalf("Find after transaction failed: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 result after transaction, got %d", len(results))
		}
	})
}

func waitForMongoDB(ctx context.Context, adapter Adapter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := adapter.Find(ctx, "test_collection", nil, 1, 0)
			if err == nil {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}
