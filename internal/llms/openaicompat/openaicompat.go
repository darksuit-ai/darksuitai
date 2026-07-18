// Package openaicompat implements a shared Chat Completions client on the
// official OpenAI Go SDK (github.com/openai/openai-go/v3). It backs both the
// OpenAI provider and the Groq provider (which is OpenAI API-compatible and only
// differs by base URL), so the two providers share one battle-tested code path
// instead of duplicating hand-rolled HTTP clients.
package openaicompat

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// Message is a provider-agnostic chat message.
type Message struct {
	Role    string
	Content string
}

// Request is a provider-agnostic Chat Completions request.
type Request struct {
	Model       string
	Messages    []Message
	MaxTokens   int
	Temperature float64
	Stop        []string
}

// Config selects the endpoint and credentials. BaseURL is empty for OpenAI and
// set to the Groq endpoint for Groq.
type Config struct {
	APIKey  string
	BaseURL string
}

func buildClient(cfg Config) (openai.Client, error) {
	if cfg.APIKey == "" {
		return openai.Client{}, fmt.Errorf("openaicompat: API key is not set")
	}
	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	return openai.NewClient(opts...), nil
}

func buildParams(req Request) openai.ChatCompletionNewParams {
	msgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages))
	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			msgs = append(msgs, openai.SystemMessage(m.Content))
		case "assistant":
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		default:
			msgs = append(msgs, openai.UserMessage(m.Content))
		}
	}
	params := openai.ChatCompletionNewParams{
		Model:    req.Model,
		Messages: msgs,
	}
	if req.MaxTokens > 0 {
		params.MaxCompletionTokens = openai.Int(int64(req.MaxTokens))
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}
	if len(req.Stop) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: req.Stop}
	}
	return params
}

// Complete performs a non-streaming Chat Completions request.
func Complete(ctx context.Context, cfg Config, req Request) (string, error) {
	client, err := buildClient(cfg)
	if err != nil {
		return "", err
	}
	resp, err := client.Chat.Completions.New(ctx, buildParams(req))
	if err != nil {
		return "", fmt.Errorf("openaicompat: request failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}

// StreamComplete streams a request and accumulates the full text.
func StreamComplete(ctx context.Context, cfg Config, req Request) (string, error) {
	client, err := buildClient(cfg)
	if err != nil {
		return "", err
	}
	stream := client.Chat.Completions.NewStreaming(ctx, buildParams(req))
	var sb strings.Builder
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			sb.WriteString(chunk.Choices[0].Delta.Content)
		}
	}
	if err := stream.Err(); err != nil {
		return sb.String(), fmt.Errorf("openaicompat: stream failed: %w", err)
	}
	return sb.String(), nil
}

// Stream streams a request, forwarding each text chunk to chunkChan. The channel
// is always closed when the function returns.
func Stream(ctx context.Context, cfg Config, req Request, chunkChan chan string) error {
	defer close(chunkChan)

	client, err := buildClient(cfg)
	if err != nil {
		chunkChan <- err.Error()
		return err
	}
	stream := client.Chat.Completions.NewStreaming(ctx, buildParams(req))
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			chunkChan <- chunk.Choices[0].Delta.Content
		}
	}
	if err := stream.Err(); err != nil {
		wrapped := fmt.Errorf("openaicompat: stream failed: %w", err)
		chunkChan <- wrapped.Error()
		return wrapped
	}
	return nil
}
