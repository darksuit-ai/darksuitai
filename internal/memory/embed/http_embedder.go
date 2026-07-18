// Package embed provides a dependency-free text embedder for semantic memory.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPEmbedder calls an OpenAI-compatible /embeddings endpoint over HTTP and
// implements memory.Embedder. It has no third-party dependencies, so it works
// with OpenAI, Azure OpenAI, Voyage (OpenAI-compatible), or any local server
// that speaks the same request/response shape.
type HTTPEmbedder struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
}

// NewHTTPEmbedder builds an embedder. Defaults: model "text-embedding-3-small",
// endpoint https://api.openai.com/v1/embeddings, 30s HTTP timeout.
func NewHTTPEmbedder(apiKey, model string) *HTTPEmbedder {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &HTTPEmbedder{
		apiKey:   apiKey,
		model:    model,
		endpoint: "https://api.openai.com/v1/embeddings",
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// WithEndpoint overrides the embeddings endpoint (e.g. Azure or a local server).
func (e *HTTPEmbedder) WithEndpoint(endpoint string) *HTTPEmbedder {
	if endpoint != "" {
		e.endpoint = endpoint
	}
	return e
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Embed returns the embedding vector for text.
func (e *HTTPEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	payload, err := json.Marshal(embedRequest{Model: e.model, Input: text})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var decoded embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("embed: decoding response: %w", err)
	}
	if decoded.Error != nil {
		return nil, fmt.Errorf("embed: API error: %s", decoded.Error.Message)
	}
	if len(decoded.Data) == 0 || len(decoded.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embed: empty embedding in response (status %d)", resp.StatusCode)
	}
	return decoded.Data[0].Embedding, nil
}
