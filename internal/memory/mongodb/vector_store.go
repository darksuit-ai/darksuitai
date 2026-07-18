package mongodb

import (
	"context"

	"github.com/darksuit-ai/darksuitai/internal/memory"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoVectorStore is a memory.VectorStore backed by MongoDB Atlas Vector
// Search. Documents store the text and its embedding; Search runs a
// $vectorSearch aggregation and returns the nearest neighbours by similarity
// score.
//
// It requires an Atlas Vector Search index on the embedding Path (default
// "embedding"). Example index definition:
//
//	{
//	  "fields": [
//	    { "type": "vector", "path": "embedding", "numDimensions": 1536, "similarity": "cosine" }
//	  ]
//	}
type MongoVectorStore struct {
	collection    *mongo.Collection
	indexName     string
	path          string
	numCandidates int
}

// NewMongoVectorStore wraps a collection and the name of its Atlas Vector Search
// index. The embedding field path defaults to "embedding".
func NewMongoVectorStore(collection *mongo.Collection, indexName string) *MongoVectorStore {
	return &MongoVectorStore{
		collection:    collection,
		indexName:     indexName,
		path:          "embedding",
		numCandidates: 100,
	}
}

// Add upserts an embedded text entry keyed by id.
func (s *MongoVectorStore) Add(ctx context.Context, id, text string, vector []float32, meta map[string]any) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"text": text, s.path: vector, "meta": meta}},
		options.Update().SetUpsert(true),
	)
	return err
}

type vectorHitDoc struct {
	ID    string         `bson:"_id"`
	Text  string         `bson:"text"`
	Score float64        `bson:"score"`
	Meta  map[string]any `bson:"meta"`
}

// Search returns the k nearest neighbours to vector via Atlas $vectorSearch.
func (s *MongoVectorStore) Search(ctx context.Context, vector []float32, k int) ([]memory.Hit, error) {
	if k <= 0 {
		k = 5
	}
	pipeline := []bson.M{
		{
			"$vectorSearch": bson.M{
				"index":         s.indexName,
				"path":          s.path,
				"queryVector":   vector,
				"numCandidates": s.numCandidates,
				"limit":         k,
			},
		},
		{
			"$project": bson.M{
				"_id":   1,
				"text":  1,
				"meta":  1,
				"score": bson.M{"$meta": "vectorSearchScore"},
			},
		},
	}

	cur, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var hits []memory.Hit
	for cur.Next(ctx) {
		var doc vectorHitDoc
		if decErr := cur.Decode(&doc); decErr != nil {
			return nil, decErr
		}
		hits = append(hits, memory.Hit{ID: doc.ID, Text: doc.Text, Score: doc.Score, Meta: doc.Meta})
	}
	if curErr := cur.Err(); curErr != nil {
		return nil, curErr
	}
	return hits, nil
}
