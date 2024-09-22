# Hapiio Database Adapter

A Go package to abstract database operations and allow for interaction with various types of databases.

## Features

- Unified interface for multiple database types
- Support for MongoDB
- Support for SQL databases (MySQL and PostgreSQL)
- CRUD operations (Create, Read, Update, Delete)
- Transaction support
- Batch insert and update operations
- Raw query execution

## Installation

To install the library, use the following command:

```sh
 go get github.com/hapiio/db-adaptor
```

## Usage

### Initializing an Adapter

```go
import (
    "github.com/hapiio/db-adapter/db"
)

// For MongoDB
config := db.Config{
    Type:             "mongodb",
    ConnectionString: "mongodb://localhost:27017",
}

// For MySQL
config := db.Config{
    Type:             "mysql",
    ConnectionString: "user:password@tcp(localhost:3306)/dbname",
    MaxOpenConns:     10,
    MaxIdleConns:     5,
}

// For PostgreSQL
config := db.Config{
    Type:             "postgres",
    ConnectionString: "postgres://user:password@localhost:5432/dbname?sslmode=disable",
    MaxOpenConns:     10,
    MaxIdleConns:     5,
}

adapter, err := db.NewAdapter(config)
if err != nil {
    // Handle error
}
defer adapter.Close()
```

### Basic Operations

```go
ctx := context.Background()

// Insert
err := adapter.Insert(ctx, "users", map[string]interface{}{
    "name": "John Doe",
    "age":  30,
})

// Find
results, err := adapter.Find(ctx, "users", map[string]interface{}{"name": "John Doe"}, 10, 0)

// Update
err := adapter.Update(ctx, "users", 
    map[string]interface{}{"age": 31}, 
    map[string]interface{}{"name": "John Doe"},
)

// Delete
err := adapter.Delete(ctx, "users", map[string]interface{}{"name": "John Doe"})
```

### Transactions

```go
ctx := context.Background()
tx, err := adapter.BeginTransaction(ctx)
if err != nil {
    // Handle error
}

// Perform operations within the transaction

err = tx.Commit()
if err != nil {
    // Handle error
}
```

### Raw Queries (SQL only)

```go
ctx := context.Background()
rows, err := adapter.QueryRaw(ctx, "SELECT * FROM users WHERE age > ?", 30)
if err != nil {
    // Handle error
}
defer rows.Close()

// Process the rows
```

## Testing

```sh
docker-compose up --build
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
