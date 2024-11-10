package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	Collections MongoCollections
)

// initializeMongoDB sets up MongoDB collections and client
func MongoChatMemory(dbName, dbConnURL string) *mongo.Collection {
	var _ = &Collections
	db, dbClient := initializeDatabase(dbConnURL, dbName)
	Collections.ChatMemory = db.Collection("darksuitchatmemory")

	Collections.Mu.Lock()
	Collections.client = dbClient
	Collections.Mu.Unlock()

	return Collections.ChatMemory
}
func initializeDatabase(dbConnectionString, dbName string) (*mongo.Database, *mongo.Client) {
	// Set up the MongoDB connection URL

	// Connect to the MongoDB database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbConnectionString))
	if err != nil {
		log.Fatal(err)
	}
	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Get a handle to the database and collections
	db := client.Database(dbName)
	_ = db.CreateCollection(ctx, "darksuitchatmemory", nil)

	return db, client
}
