package openai

import (
	"context"

	"github.com/darksuit-ai/darksuitai/internal/llms/openai/types"
	"github.com/darksuit-ai/darksuitai/internal/llms/openaicompat"
)

// This file was migrated from a hand-rolled net/http client to the official
// OpenAI Go SDK via the shared openaicompat package (Chat Completions API). The
// custom RateLimiter, transport, retry loop and SSE scanner are gone; the
// package's public surface (Client, StreamCompleteClient, StreamClient and the
// OAIChatArgs methods in api.go) is unchanged.
//
// Note: this uses the Chat Completions API (shared with the Groq provider),
// which OpenAI fully supports. Migrating to the newer Responses API is a
// possible follow-up if server-side tools/state are desired.

func toRequest(req types.ChatArgs) openaicompat.Request {
	msgs := make([]openaicompat.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, openaicompat.Message{Role: m.Role, Content: m.Content})
	}
	return openaicompat.Request{
		Model:       req.Model,
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stop:        toStringSlice(req.Stop),
	}
}

func toStringSlice(v interface{}) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []interface{}:
		out := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

func config(apiKey string) openaicompat.Config {
	// Empty BaseURL => the SDK's default OpenAI endpoint.
	return openaicompat.Config{APIKey: apiKey}
}

// Client performs a non-streaming chat completion.
func Client(apiKey string, req types.ChatArgs) (string, error) {
	return openaicompat.Complete(context.Background(), config(apiKey), toRequest(req))
}

// StreamCompleteClient streams a chat completion and returns the full text.
func StreamCompleteClient(apiKey string, req types.ChatArgs) (string, error) {
	return openaicompat.StreamComplete(context.Background(), config(apiKey), toRequest(req))
}

// StreamClient streams a chat completion, forwarding chunks to chunkChan.
func StreamClient(apiKey string, req types.ChatArgs, chunkChan chan string) error {
	return openaicompat.Stream(context.Background(), config(apiKey), toRequest(req), chunkChan)
}
