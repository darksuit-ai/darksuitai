package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/darksuit-ai/darksuitai/internal/llms/anthropic/types"
)

// RateLimiter is a simple rate limiter implementation
type RateLimiter struct {
	sync.Mutex
	lastRequest time.Time
	maxRate     time.Duration
}

// Wait blocks until it's time to allow the next request
func (r *RateLimiter) Wait() {
	r.Lock()
	defer r.Unlock()
	now := time.Now()
	if elapsed := now.Sub(r.lastRequest); elapsed < r.maxRate {
		time.Sleep(r.maxRate - elapsed)
	}
	r.lastRequest = now
}

// rateLimiter is a global rate limiter instance
var rateLimiter = RateLimiter{
	maxRate: 1 * time.Second, // Adjust the rate limit as needed
}

// Configure the HTTP transport for connection reuse
var transport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 10 * time.Second,
}

// Global HTTP client to reuse across requests
var httpClient = &http.Client{
	Transport: transport,
	Timeout:   0, // No timeout for streaming; use context for control
}



func retryRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 5; i++ { // Retry up to 5 times
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		timeout := calculateRetryTimeout(i)
		time.Sleep(timeout)
	}
	return resp, err
}

func calculateRetryTimeout(retryCount int) time.Duration {
	// Exponential backoff with jitter
	sleepSeconds := math.Min(float64(int(1<<retryCount)), 60) // Cap at 60 seconds
	jitter := sleepSeconds * (1 + 0.25*(rand.Float64()-0.5))
	return time.Duration(jitter) * time.Second
}

func checkEnvVar(anthropicAPIKey string) {
	// Check for the required environment variable
	if anthropicAPIKey == "" {
		log.Fatal(`
ANTHROPIC_API_KEY is not set or is empty. 
Please set it in the .env file as follows:

    ANTHROPIC_API_KEY="your_anthropic_api_key_here"

Make sure to replace "your_anthropic_api_key_here" with your actual Anthropic API key.
If you don't have a .env file, create one in the root directory of your project.
`)
	}
}

func Client(apiKey string, req types.ChatArgs) (string, error) {
	checkEnvVar(apiKey)
	// Wait for rate limiter
	//rateLimiter.Wait()
	// Marshal the payload to JSON
	reqJsonPayload, err := json.Marshal(req)

	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer([]byte(reqJsonPayload)))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	request.Header.Set("x-api-key", apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("anthropic-version", "2023-06-01")
	request.Header.Set("anthropic-beta", "messages-2023-12-15")

	// Make the request with retry logic
	resp, err := retryRequest(httpClient, request)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	// print(resp.StatusCode)
	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading response body: %w", err)
		}

		// Convert the response body to a string
		bodyString := string(bodyBytes)

		// Print the response body, neccessary for immediate error
		fmt.Println(bodyString)
		// Print the response body
		return bodyString, nil
	}
	// Check if the response status indicates an error
	if resp.StatusCode >= 400 {
		var clientErr types.ChatError
		if err := json.NewDecoder(resp.Body).Decode(&clientErr); err != nil {
			return "", fmt.Errorf("error unmarshaling error response: %w", err)
		}
		return "", fmt.Errorf("API error: %v", clientErr)
	}

	// Unmarshal the successful response
	var response types.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error unmarshaling chat completion response: %w", err)
	}
	// Extract the content from the response
	content := response.Content[0]["text"]

	return content, nil
}

func StreamCompleteClient(apiKey string, req types.ChatArgs) (string, error) {
	checkEnvVar(apiKey)
	// Marshal the payload to JSON
	reqJsonPayload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer([]byte(reqJsonPayload)))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("x-api-key", apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")
	request.Header.Set("anthropic-beta", "messages-2023-12-15")

	// Make the request
	resp, err := retryRequest(httpClient, request)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading response body: %w", err)
		}

		// Convert the response body to a string
		bodyString := string(bodyBytes)

		// Print the response body
		fmt.Println(bodyString)
	}
	// Check if the response status indicates an error
	if resp.StatusCode >= 400 {
		var clientErr types.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&clientErr); err != nil {
			return "", fmt.Errorf("error unmarshaling error response: %w", err)
		}
		return "", fmt.Errorf("API error: %v", clientErr)
	}
	// Use a scanner to read the streaming response
	scanner := bufio.NewScanner(resp.Body)
	result := []string{}

	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, []byte(`event: message_stop)`)) {
			break
		}
		if bytes.HasPrefix(line, []byte(`data: `)) {
			data := bytes.TrimPrefix(line, []byte(`data: `))
			var chunk types.ContentBlockDelta
			if err := json.Unmarshal(data, &chunk); err != nil {
				result = append(result, err.Error())
			}
			if chunk.Type == "ping" {
				continue
			}
			if chunk.Type == "content_block_start" {
				continue
			}
			if chunk.Type == "content_block_delta" {
				result = append(result, chunk.Delta["text"])
			}
		} else if bytes.HasPrefix(line, []byte(`event: error)`)) {
			data := bytes.TrimPrefix(line, []byte(`data: `))
			var chunk types.ChatOverloadError
			if err := json.Unmarshal(data, &chunk); err != nil {
				result = append(result, err.Error())
			}
		}
	}
	finalResult := strings.Join(result, "")
	return finalResult, nil
}

func StreamClient(apiKey string, req types.ChatArgs, chunkChan chan string) error {
	// Ensure we close the channel when we're done
	defer close(chunkChan)
	
	checkEnvVar(apiKey)
	// Marshal the payload to JSON
	reqJsonPayload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer([]byte(reqJsonPayload)))
	if err != nil {
		return err
	}

	// Set request headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("x-api-key", apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")
	request.Header.Set("anthropic-beta", "messages-2023-12-15")

	// Make the request
	resp, err := retryRequest(httpClient, request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Convert the response body to a string
		bodyString := string(bodyBytes)

		// Print the response body
		fmt.Println(bodyString)
	}
	// Check if the response status indicates an error
	if resp.StatusCode >= 400 {
		var clientErr types.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&clientErr); err != nil {
			return err
		}

		errorMessage := clientErr.Error.Message
		fmt.Println(errorMessage)

		chunkChan <- errorMessage

	}
	// Use a scanner to read the streaming response
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, []byte(`event: message_stop)`)) {
			break
		}
		if bytes.HasPrefix(line, []byte(`data: `)) {
			data := bytes.TrimPrefix(line, []byte(`data: `))
			var chunk types.ContentBlockDelta
			if err := json.Unmarshal(data, &chunk); err != nil {
				chunkChan <- err.Error()
			}
			if chunk.Type == "ping" {
				continue
			}
			if chunk.Type == "content_block_start" {
				continue
			}
			if chunk.Type == "content_block_delta" {
				chunkChan <- chunk.Delta["text"]

			}
		} else if bytes.HasPrefix(line, []byte(`event: error)`)) {
			data := bytes.TrimPrefix(line, []byte(`data: `))
			var chunk types.ChatOverloadError
			if err := json.Unmarshal(data, &chunk); err != nil {
				chunkChan <- err.Error()
			}
		}
	}
	return nil
}
