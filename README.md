# Database Adapter

A Go package to abstract database operations and allow for interaction with various types of databases.

## Supported Databases

- MongoDB
- MySQL
- PostgreSQL

## Getting Started

These instructions will get you a copy of the project up and running on your local machine.

### Prerequisites

- Go (1.16+)
- MongoDB/SQL Database running and accessible

### Installing

1. Clone the db-adapter repository into your project:

   ```sh
    go get github.com/hapiio/db-adapter
    ```

2. Import db-adapter in your Go files:

```sh
import "github.com/yourusername/db-adapter/dbadapter"
```

### Usage

Here's a quick example of how to use the database adapter with a MongoDB database:

```go
package main

import (
    "context"
    "log"
    "github.com/hapiio/db-adapter"
)

func main() {
    ctx := context.Background()
    
    mongoAdapter := dbadapter.NewMongoDBAdapter()
    if err := mongoAdapter.Connect(ctx, "your-mongodb-connection-string"); err != nil {
        log.Fatal(err)
    }
    defer mongoAdapter.Close()

    // Now you can use mongoAdapter to interact with your MongoDB database.
}

```

And here's an example with an SQL database:

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "github.com/hapiio/db-adapter"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    ctx := context.Background()

    sqlDB, err := sql.Open("mysql", "your-dsn")
    if err != nil {
        log.Fatal(err)
    }
    defer sqlDB.Close()

    sqlAdapter := dbadapter.NewSQLAdapter(sqlDB)
    // Now you can use sqlAdapter to interact with your SQL database.
}
```

Replace `your-mongodb-connection-string` and `your-dsn` with your actual database connection strings.
