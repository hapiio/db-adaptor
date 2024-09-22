package db

import (
	"context"
	"database/sql"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBAdapter struct {
	client   *mongo.Client
	database *mongo.Database
}

func (m *MongoDBAdapter) Connect(config Config) error {
	clientOptions := options.Client().ApplyURI(config.ConnectionString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}
	m.client = client
	m.database = client.Database("your_database_name")
	return nil
}

func (m *MongoDBAdapter) Close() error {
	if m.client != nil {
		return m.client.Disconnect(context.TODO())
	}
	return nil
}

func (m *MongoDBAdapter) Insert(ctx context.Context, collection string, document map[string]interface{}) error {
	coll := m.database.Collection(collection)
	_, err := coll.InsertOne(ctx, document)
	return err
}

func (m *MongoDBAdapter) Update(ctx context.Context, collection string, document map[string]interface{}, filter map[string]interface{}) error {
	coll := m.database.Collection(collection)
	_, err := coll.UpdateOne(ctx, filter, bson.M{"$set": document})
	return err
}

func (m *MongoDBAdapter) Delete(ctx context.Context, collection string, filter map[string]interface{}) error {
	coll := m.database.Collection(collection)
	_, err := coll.DeleteOne(ctx, filter)
	return err
}

func (m *MongoDBAdapter) Find(ctx context.Context, collection string, filter map[string]interface{}, limit, offset int) ([]map[string]interface{}, error) {
	coll := m.database.Collection(collection)
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (m *MongoDBAdapter) ExecuteRaw(ctx context.Context, command string, args ...interface{}) (sql.Result, error) {
	return nil, fmt.Errorf("ExecuteRaw is not supported for MongoDB")
}

func (m *MongoDBAdapter) QueryRaw(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("QueryRaw is not supported for MongoDB")
}

func (m *MongoDBAdapter) BeginTransaction(ctx context.Context) (Transaction, error) {
	session, err := m.client.StartSession()
	if err != nil {
		return nil, err
	}
	if err = session.StartTransaction(); err != nil {
		return nil, err
	}
	return &MongoTransaction{session: session}, nil
}

type MongoTransaction struct {
	session mongo.Session
}

func (t *MongoTransaction) Commit() error {
	return t.session.CommitTransaction(context.Background())
}

func (t *MongoTransaction) Rollback() error {
	return t.session.AbortTransaction(context.Background())
}
