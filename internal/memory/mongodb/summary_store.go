package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoSummaryStore persists a session's rolling conversation summary. It
// implements memory.SummaryStore (Phase 4 compaction) without importing the
// memory package, avoiding an import cycle.
type MongoSummaryStore struct {
	collection *mongo.Collection
}

// NewMongoSummaryStore wraps a collection used to store rolling summaries.
func NewMongoSummaryStore(collection *mongo.Collection) *MongoSummaryStore {
	return &MongoSummaryStore{collection: collection}
}

type summaryDoc struct {
	SessionID     string `bson:"sessionId" json:"sessionId"`
	Summary       string `bson:"summary" json:"summary"`
	CompactedUpTo int    `bson:"compactedUpTo" json:"compactedUpTo"`
}

// GetSummary returns the stored summary and coverage for a session. A missing
// session yields ("", 0, nil).
func (s *MongoSummaryStore) GetSummary(ctx context.Context, sessionID string) (string, int, error) {
	var doc summaryDoc
	err := s.collection.FindOne(ctx, bson.M{"sessionId": sessionID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", 0, nil
		}
		return "", 0, err
	}
	return doc.Summary, doc.CompactedUpTo, nil
}

// SetSummary upserts the rolling summary and coverage for a session.
func (s *MongoSummaryStore) SetSummary(ctx context.Context, sessionID, summary string, compactedUpTo int) error {
	_, err := s.collection.UpdateOne(ctx,
		bson.M{"sessionId": sessionID},
		bson.M{"$set": bson.M{"summary": summary, "compactedUpTo": compactedUpTo}},
		options.Update().SetUpsert(true),
	)
	return err
}
