package anthropic

import (
	"context"
	"fmt"
	"strings"

	"github.com/darksuit-ai/darksuitai/internal/llms/anthropic/types"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// This file was migrated (Phase 1) from a hand-rolled net/http client to the
// official Anthropic Go SDK (github.com/anthropics/anthropic-sdk-go).
//
// The SDK provides retries, connection pooling, streaming (SSE) decoding and
// typed request/response models out of the box, so the previous RateLimiter,
// custom http.Transport, manual retry/backoff and hand-written SSE scanner have
// all been removed. The public surface of this package (Client,
// StreamCompleteClient, StreamClient and the AnthChatArgs methods in api.go) is
// unchanged, so callers in pkg/agent, pkg/chat and pkg/convchat need no edits.

// newSDKClient builds an Anthropic SDK client from an explicit API key.
// The key is passed through the existing framework configuration rather than
// relying solely on the ANTHROPIC_API_KEY environment variable.
func newSDKClient(apiKey string) (anthropic.Client, error) {
	if apiKey == "" {
		return anthropic.Client{}, fmt.Errorf("anthropic: API key is not set; pass it via AddAPIKey or the ANTHROPIC_API_KEY environment variable")
	}
	return anthropic.NewClient(option.WithAPIKey(apiKey)), nil
}

// buildParams converts the framework's provider-agnostic ChatArgs into the
// SDK's MessageNewParams.
func buildParams(req types.ChatArgs) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: int64(req.MaxTokens),
		Messages:  toSDKMessages(req.Messages),
	}

	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}
	// NOTE: temperature is intentionally NOT sent. Anthropic's 2026 models
	// (e.g. claude-sonnet-5, claude-opus-4.x) deprecated the `temperature`
	// parameter and reject requests that include it ("`temperature` is
	// deprecated for this model"). Sampling is managed by the model. Older
	// models that accepted temperature still work without it.
	if stops := toStringSlice(req.StopSequences); len(stops) > 0 {
		params.StopSequences = stops
	}
	return params
}

// toSDKMessages maps the framework's simple role/content messages onto the
// SDK's MessageParam blocks. Anthropic only supports user/assistant turns, so
// any other role is treated as a user turn.
func toSDKMessages(messages []types.Message) []anthropic.MessageParam {
	out := make([]anthropic.MessageParam, 0, len(messages))
	for _, m := range messages {
		block := anthropic.NewTextBlock(m.Content)
		switch m.Role {
		case "assistant":
			out = append(out, anthropic.NewAssistantMessage(block))
		default:
			out = append(out, anthropic.NewUserMessage(block))
		}
	}
	return out
}

// toStringSlice normalizes the loosely-typed StopSequences field (kept as
// interface{} for backwards compatibility) into a concrete []string.
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

// Client performs a non-streaming message request and returns the concatenated
// text content of the response.
func Client(apiKey string, req types.ChatArgs) (string, error) {
	client, err := newSDKClient(apiKey)
	if err != nil {
		return "", err
	}

	msg, err := client.Messages.New(context.Background(), buildParams(req))
	if err != nil {
		return "", fmt.Errorf("anthropic: message request failed: %w", err)
	}

	var sb strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}

// StreamCompleteClient streams a message request but accumulates the full text
// before returning it (the caller wants a single string, not incremental chunks).
func StreamCompleteClient(apiKey string, req types.ChatArgs) (string, error) {
	client, err := newSDKClient(apiKey)
	if err != nil {
		return "", err
	}

	stream := client.Messages.NewStreaming(context.Background(), buildParams(req))
	var sb strings.Builder
	for stream.Next() {
		event := stream.Current()
		switch eventVariant := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := eventVariant.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				sb.WriteString(delta.Text)
			}
		}
	}
	if err := stream.Err(); err != nil {
		return sb.String(), fmt.Errorf("anthropic: stream failed: %w", err)
	}
	return sb.String(), nil
}

// StreamClient streams a message request and forwards each text delta to
// chunkChan. The channel is always closed when the function returns.
func StreamClient(apiKey string, req types.ChatArgs, chunkChan chan string) error {
	defer close(chunkChan)

	client, err := newSDKClient(apiKey)
	if err != nil {
		chunkChan <- err.Error()
		return err
	}

	stream := client.Messages.NewStreaming(context.Background(), buildParams(req))
	for stream.Next() {
		event := stream.Current()
		switch eventVariant := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := eventVariant.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				chunkChan <- delta.Text
			}
		}
	}
	if err := stream.Err(); err != nil {
		wrapped := fmt.Errorf("anthropic: stream failed: %w", err)
		chunkChan <- wrapped.Error()
		return wrapped
	}
	return nil
}
