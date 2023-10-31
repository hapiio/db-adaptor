package main

import (
	"log"

	db "github.com/hapiio/db-adaptor/db"
)

func main() {
	// MongoDB Adapter Example
	mongoDB := &db.MongoDBAdapter{}
	if err := mongoDB.Connect("your_mongodb_connection_string"); err != nil {
		log.Fatal(err)
	}
	defer mongoDB.Close()

	document := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	if err := mongoDB.Insert("your_collection", document); err != nil {
		log.Fatal(err)
	}

	// SQL Adapter Example
	sqlDB := &db.SQLAdapter{}
	if err := sqlDB.Connect("your_mysql_connection_string", "mysql"); err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Insert("INSERT INTO your_table (column1, column2) VALUES (?, ?)", "value1", "value2"); err != nil {
		log.Fatal(err)
	}
}
