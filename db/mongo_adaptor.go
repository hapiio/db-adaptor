package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBAdapter struct {
	client   *mongo.Client
	database *mongo.Database
}

func (m *MongoDBAdapter) Connect(connectionString string) error {
	clientOptions := options.Client().ApplyURI(connectionString)
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

func (m *MongoDBAdapter) Insert(collectionName string, document interface{}) error {
	collection := m.database.Collection(collectionName)
	_, err := collection.InsertOne(context.TODO(), document)
	return err
}

func (m *MongoDBAdapter) Update(ctx context.Context, collectionName string, filter interface{}, update interface{}) error {
	collection := m.database.Collection(collectionName)
	_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	return err
}

func (m *MongoDBAdapter) Delete(ctx context.Context, collectionName string, filter interface{}) error {
	collection := m.database.Collection(collectionName)
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

func (m *MongoDBAdapter) Find(ctx context.Context, collectionName string, filter interface{}) (*mongo.Cursor, error) {
	collection := m.database.Collection(collectionName)
	return collection.Find(ctx, filter)
}

func (m *MongoDBAdapter) BatchInsert(ctx context.Context, collectionName string, documents []interface{}) error {
	collection := m.database.Collection(collectionName)
	_, err := collection.InsertMany(ctx, documents)
	return err
}

func (m *MongoDBAdapter) Close() error {
	if m.client != nil {
		return m.client.Disconnect(context.TODO())
	}
	return nil
}
