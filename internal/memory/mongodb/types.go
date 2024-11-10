package mongodb

import (
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCollections struct {
	ChatMemory *mongo.Collection
	Mu         sync.RWMutex
	client     *mongo.Client
}