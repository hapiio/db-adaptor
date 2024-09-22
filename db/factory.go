package db

import "fmt"

func NewAdapter(config Config) (Adapter, error) {
	switch config.Type {
	case "mongodb":
		adapter := &MongoDBAdapter{}
		err := adapter.Connect(config)
		return adapter, err
	case "mysql", "postgres":
		adapter := &SQLAdapter{}
		err := adapter.Connect(config)
		return adapter, err
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
