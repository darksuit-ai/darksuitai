package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/darksuit-ai/darksuitai/internal/memory"
	utl "github.com/darksuit-ai/darksuitai/internal/utilities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChatMemoryCollectionInterface defines the required methods for managing chat memory operations
type ChatMemoryCollectionInterface interface {
	AddConversationToMemory(sessionId, prompt, aiMessage string) error
	RetrieveMemoryWithK(sessionId string, k int64) (string, error)
	// RetrieveTurns returns the full, chronological (oldest-first) turn history
	// for a session; used by the Phase 4 compactor.
	RetrieveTurns(sessionId string) ([]memory.Turn, error)
}

type dataObject struct {
	UserPrompt       string `bson:"user_prompt" json:"user_prompt"`
	DarksuitResponse string `bson:"darksuit_response" json:"darksuit_response"`
	ToolUsed         string `bson:"tool_used,omitempty" json:"tool_used,omitempty"`
}

type convData struct {
	Type string     `bson:"type" json:"type"`
	Data dataObject `bson:"data" json:"data"`
}

type convHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	SessionId string             `bson:"sessionId" json:"sessionId"`
	History   convData           `bson:"History" json:"History"`
	TimeStamp string             `json:"timestamp"`
}

// MongoCollection implements ChatMemoryCollectionInterface
type MongoCollection struct {
	collection *mongo.Collection
}

// NewMongoCollection creates a new MongoCollection instance
func NewMongoCollection(collection *mongo.Collection) *MongoCollection {
	return &MongoCollection{
		collection: collection,
	}
}

/*
This function creates a new memory entry in a MongoDB collection.

Args:

	userId (str): The user ID to associate with the memory.
	prompt (str): The user's prompt.
	ai_message (str): The AI's response message.

The function trims off a specified string from the ai_message before storing it.
*/
func (mc *MongoCollection) AddConversationToMemory(sessionId, prompt, aiMessage string) error {

	// Create a new convHistory struct with the provided data
	history := convHistory{
		SessionId: sessionId,
		History: convData{
			Type: "ai",
			Data: dataObject{
				UserPrompt:       prompt,
				DarksuitResponse: aiMessage,
			},
		},
		TimeStamp: time.Now().UTC().Format(time.RFC3339), // Get the current timestamp in RFC3339 format
	}

	// Insert the data into the collection
	_, dbErr := mc.collection.InsertOne(context.Background(), history)
	if dbErr != nil {
		return dbErr
	}
	return nil
}

/*
This function retrieves the most recent entries from a MongoDB collection where the userId matches the provided userId.
It only returns entries if they were created within the last 1 minute. If no matching entries are found, or if the most recent
matching entries are older than 1 minute, the function returns None.

Args:

	collection: The MongoDB collection to retrieve the memory from.
	userId (str): The user ID to match.
	k (int): The number of memory messages to pull

Returns:

	list: The most recent matching entries in the collection, or None if no match is found or if the entries are older than 1 minute.
*/
func (mc *MongoCollection) RetrieveMemoryWithK(sessionId string, k int64) (string, error) {

	// Query the collection for entries where the sessionId matches and the timestamp is within the last 15 minute
	query := bson.M{"sessionId": sessionId}

	// Find the documents, sort them by timestamp in descending order, and limit to 'k' results
	opts := options.Find().SetSort(bson.D{primitive.E{Key: "timestamp", Value: -1}}).SetLimit(k)
	// Sort the results in descending order by timestamp and retrieve the first k results
	cur, dbErr := mc.collection.Find(context.Background(), query, opts)
	if dbErr != nil {
		return "", dbErr
	}
	// Close the MongoDB cursor after iterating over the results
	defer cur.Close(context.Background())

	// The cursor returns documents newest-first (timestamp descending).
	// Collect them, then emit oldest-first so the rendered history reads
	// chronologically, with each turn ordered Human -> AI.
	//
	// NOTE: the previous implementation called sort.Slice with
	// `func(i, j int) bool { return i > j }`, which sorts by slice *index*
	// rather than by timestamp. That only reversed the slice and scrambled
	// the Human/AI pairing; it did not order the history by time.
	var docs []convHistory
	for cur.Next(context.Background()) {
		var doc convHistory
		if decErr := cur.Decode(&doc); decErr != nil {
			return "", decErr
		}
		docs = append(docs, doc)
	}
	if curErr := cur.Err(); curErr != nil {
		return "", curErr
	}

	// If no history was found, return an empty string
	if len(docs) == 0 {
		return "[]", nil
	}

	// Walk newest-first docs in reverse to build an oldest-first transcript.
	var chatHistory []string
	for i := len(docs) - 1; i >= 0; i-- {
		doc := docs[i]
		chatHistory = append(chatHistory, utl.ConcatWords([]byte(``), []byte(`Human: `), []byte(doc.History.Data.UserPrompt)))
		chatHistory = append(chatHistory, utl.ConcatWords([]byte(``), []byte(`AI: `), []byte(doc.History.Data.DarksuitResponse)))
	}

	// Join the chat_history slice elements into a single string separated by ", "
	chatHistoryString := utl.ConcatWords([]byte(""), []byte(strings.Join(chatHistory, "\n")))

	// Return the chat_history_string
	return chatHistoryString, nil

}

// RetrieveTurns returns the full conversation for a session in chronological
// (oldest-first) order, for use by the Phase 4 compactor.
func (mc *MongoCollection) RetrieveTurns(sessionId string) ([]memory.Turn, error) {
	// Ascending timestamp => oldest first.
	opts := options.Find().SetSort(bson.D{primitive.E{Key: "timestamp", Value: 1}})
	cur, dbErr := mc.collection.Find(context.Background(), bson.M{"sessionId": sessionId}, opts)
	if dbErr != nil {
		return nil, dbErr
	}
	defer cur.Close(context.Background())

	var turns []memory.Turn
	for cur.Next(context.Background()) {
		var doc convHistory
		if decErr := cur.Decode(&doc); decErr != nil {
			return nil, decErr
		}
		turns = append(turns, memory.Turn{
			Human: doc.History.Data.UserPrompt,
			AI:    doc.History.Data.DarksuitResponse,
		})
	}
	if curErr := cur.Err(); curErr != nil {
		return nil, curErr
	}
	return turns, nil
}
