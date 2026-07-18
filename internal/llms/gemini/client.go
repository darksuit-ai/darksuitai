package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/darksuit-ai/darksuitai/internal/llms/gemini/types"

	"google.golang.org/genai"
)

// This file was migrated from a hand-rolled net/http client (against Gemini's
// OpenAI-compatible endpoint) to the official Google Gen AI SDK
// (google.golang.org/genai). The SDK handles auth, retries, streaming and typed
// models, so the previous RateLimiter, custom transport, manual retry loop and
// SSE scanner have been removed. The package's public surface (Client,
// StreamCompleteClient, StreamClient and the GEMChatArgs methods in api.go) is
// unchanged, so callers need no edits.

const defaultGeminiModel = "gemini-2.5-flash"

// newGenaiClient builds a Gemini SDK client from an explicit API key.
func newGenaiClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini: API key is not set; pass it via AddAPIKey or the GEMINI_API_KEY environment variable")
	}
	return genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
}

// splitPromptAndSystem derives the user prompt and system instruction from the
// framework's provider-agnostic message list. Gemini takes the system prompt
// out-of-band (SystemInstruction), so system-role messages are separated out.
func splitPromptAndSystem(messages []types.Message) (prompt, system string) {
	var b strings.Builder
	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		b.WriteString(m.Content)
	}
	return b.String(), system
}

// buildConfig maps ChatArgs onto the SDK's generation config.
func buildConfig(req types.ChatArgs, system string) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{}
	if req.Temperature > 0 {
		t := float32(req.Temperature)
		cfg.Temperature = &t
	}
	if req.MaxTokens > 0 {
		cfg.MaxOutputTokens = int32(req.MaxTokens)
	}
	if stops := toStringSlice(req.Stop); len(stops) > 0 {
		cfg.StopSequences = stops
	}
	if system != "" {
		cfg.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: system}}}
	}
	return cfg
}

// toStringSlice normalizes the loosely-typed Stop field into a concrete []string.
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

func modelOrDefault(model string) string {
	if model == "" {
		return defaultGeminiModel
	}
	return model
}

// Client performs a non-streaming generate request and returns the text.
func Client(apiKey string, req types.ChatArgs) (string, error) {
	ctx := context.Background()
	client, err := newGenaiClient(ctx, apiKey)
	if err != nil {
		return "", err
	}

	prompt, system := splitPromptAndSystem(req.Messages)
	resp, err := client.Models.GenerateContent(ctx, modelOrDefault(req.Model), genai.Text(prompt), buildConfig(req, system))
	if err != nil {
		return "", fmt.Errorf("gemini: generate failed: %w", err)
	}
	return resp.Text(), nil
}

// StreamCompleteClient streams a generate request and accumulates the full text.
func StreamCompleteClient(apiKey string, req types.ChatArgs) (string, error) {
	ctx := context.Background()
	client, err := newGenaiClient(ctx, apiKey)
	if err != nil {
		return "", err
	}

	prompt, system := splitPromptAndSystem(req.Messages)
	var sb strings.Builder
	var streamErr error
	// GenerateContentStream returns an iter.Seq2 (a func); call it directly with
	// a yield callback so this compiles across Go toolchains.
	stream := client.Models.GenerateContentStream(ctx, modelOrDefault(req.Model), genai.Text(prompt), buildConfig(req, system))
	stream(func(result *genai.GenerateContentResponse, err error) bool {
		if err != nil {
			streamErr = err
			return false
		}
		sb.WriteString(result.Text())
		return true
	})
	if streamErr != nil {
		return sb.String(), fmt.Errorf("gemini: stream failed: %w", streamErr)
	}
	return sb.String(), nil
}

// StreamClient streams a generate request, forwarding each text chunk to
// chunkChan. The channel is always closed when the function returns.
func StreamClient(apiKey string, req types.ChatArgs, chunkChan chan string) error {
	defer close(chunkChan)

	ctx := context.Background()
	client, err := newGenaiClient(ctx, apiKey)
	if err != nil {
		chunkChan <- err.Error()
		return err
	}

	prompt, system := splitPromptAndSystem(req.Messages)
	var streamErr error
	stream := client.Models.GenerateContentStream(ctx, modelOrDefault(req.Model), genai.Text(prompt), buildConfig(req, system))
	stream(func(result *genai.GenerateContentResponse, err error) bool {
		if err != nil {
			streamErr = err
			return false
		}
		chunkChan <- result.Text()
		return true
	})
	if streamErr != nil {
		wrapped := fmt.Errorf("gemini: stream failed: %w", streamErr)
		chunkChan <- wrapped.Error()
		return wrapped
	}
	return nil
}
