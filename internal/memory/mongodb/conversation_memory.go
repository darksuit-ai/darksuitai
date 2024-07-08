package mongodb

import (
	"context"
	exp "github.com/darksuit-ai/darksuitai/internal/exceptions"
	utl "github.com/darksuit-ai/darksuitai/internal/utilities"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dataObject struct {
	UserPrompt string `bson:"user_prompt" json:"user_prompt"`
	AIResponse string `bson:"ai_response" json:"ai_response"`
}

type convData struct {
	Type string     `bson:"type" json:"type"`
	Data dataObject `bson:"data" json:"data"`
}

type convHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserId    string             `bson:"userId" json:"userId"`
	History   convData           `bson:"History" json:"History"`
	TimeStamp string             `json:"timestamp"`
}

/*
This function creates a new memory entry in a MongoDB collection for a specified user.

Args:

	userId (str): The user ID to associate with the memory.
	prompt (str): The user's prompt.
	ai_message (str): The AI's response message.

The function trims off a specified string from the ai_message before storing it.
*/
func AddConversationToMemory(db *mongo.Database, userId string, prompt string, ai_message string, toolsResponse string) {
	var chatMemory = db.Collection("darksuitaichatmemory")
	timeStamp := time.Now()
	// Create a new convHistory struct with the provided data
	history := convHistory{
		UserId: userId,
		History: convData{
			Type: "ai",
			Data: dataObject{
				UserPrompt: prompt,
				AIResponse: ai_message,
			},
		},
		TimeStamp: timeStamp.Format("2006-01-02T15:04:05.000000"), // Get the current timestamp in RFC3339 format
	}

	// Insert the data into the collection
	_, dbErr := chatMemory.InsertOne(context.Background(), history)
	if dbErr != nil {
		exp.Loggers.System.Warn(dbErr)
	}
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
func RetrieveMemoryWithK(db *mongo.Database, userId string, k int64) string {
	var chatMemory = db.Collection("darksuitaichatmemory")
	// Query the collection for entries where the userId matches and the timestamp is within the last 15 minute
	//minTime := time.Now().Add(-15 * time.Minute)
	// Format the time to the desired output format
	//formattedTime := minTime.Format("2006-01-02T15:04:05.000000")

	// Construct a MongoDB query to find documents where the userId matches and the timestamp is greater than or equal to the formattedTime
	//query := bson.M{"userId": userId, "timestamp": bson.M{"$gte": formattedTime}}
	query := bson.M{"userId": userId}

	// Find the documents, sort them by timestamp in ascending order, and limit to 'k' results
	// IMPORTANT keep value as -1
	opts := options.Find().SetSort(bson.D{primitive.E{Key: "timestamp", Value: -1}}).SetLimit(k)

	// Sort the results in descending order by timestamp and retrieve the first k results
	cur, dbErr := chatMemory.Find(context.Background(), query, opts)
	if dbErr != nil {
		exp.Loggers.System.Warn(dbErr)
	}
	// Close the MongoDB cursor after iterating over the results
	defer cur.Close(context.Background())

	// Initialize a slice to store the chat history
	var chatHistory []string

	// Iterate over the cursor and append the user prompt and AI response to the chat_history slice
	for cur.Next(context.Background()) {
		var doc convHistory
		err := cur.Decode(&doc)
		if err != nil {
			exp.Loggers.System.Warn(dbErr)
		}

		chatHistory = append(chatHistory, utl.ConcatWords([]byte(``), []byte(`AIMessage=(`), []byte(doc.History.Data.AIResponse), []byte(`)`)))
		chatHistory = append(chatHistory, utl.ConcatWords([]byte(``), []byte(`HumanMessage=(`), []byte(doc.History.Data.UserPrompt), []byte(`)`)))
	}

	// reverse the array
	sort.Slice(chatHistory, func(i, j int) bool {
		return i > j
	})

	// If the chat_history slice is empty, return an empty string
	if len(chatHistory) == 0 {
		return "[]"
	}

	// Join the chat_history slice elements into a single string separated by ", "
	chatHistoryString := utl.ConcatWords([]byte(""), []byte("["), []byte(strings.Join(chatHistory, ", ")), []byte("]"))

	// Return the chat_history_string
	return chatHistoryString

}
