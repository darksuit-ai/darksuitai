// Package observability provides a lightweight, dependency-free tracing seam
// for the agent runtime (Phase 3).
//
// The design is span-oriented: an Observer produces a RunHandle for each agent
// invocation, and all events for that invocation are recorded on its own
// handle. This keeps concurrent agent runs isolated without shared mutable
// state.
//
// Event field names deliberately mirror the OpenTelemetry GenAI semantic
// conventions (gen_ai.system, gen_ai.request.model, gen_ai.usage.*,
// gen_ai.operation.name, etc.) so an OpenTelemetry-backed Observer can map them
// one-to-one. Built-in observers (no-op, stdout JSON, LangSmith REST) require no
// external dependencies; an OpenTelemetry adapter can be supplied by the
// application by implementing Observer (see docs/PHASE3_OBSERVABILITY.md).
package observability

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// RunInfo describes a single agent invocation.
type RunInfo struct {
	SessionID string
	Provider  string // gen_ai.system (e.g. "anthropic")
	Model     string // gen_ai.request.model
	Protocol  string // "xml" or "native"
	Input     string
}

// LLMCall captures a single model call's token usage.
type LLMCall struct {
	Provider     string
	Model        string
	InputTokens  int // gen_ai.usage.input_tokens
	OutputTokens int // gen_ai.usage.output_tokens
}

// ToolCall captures a single tool execution.
type ToolCall struct {
	Name     string
	Input    string
	Output   string
	IsError  bool
	Duration time.Duration
}

// Observer starts a RunHandle for each agent invocation.
type Observer interface {
	StartRun(info RunInfo) RunHandle
}

// RunHandle records events for one agent invocation and is closed with End.
type RunHandle interface {
	LLMEnd(call LLMCall)
	ToolEnd(call ToolCall)
	Iteration(n int)
	Error(stage string, err error)
	End(output string, err error)
}

// ---- no-op ----

// Noop is the default Observer; it records nothing.
type Noop struct{}

type noopHandle struct{}

func (Noop) StartRun(RunInfo) RunHandle { return noopHandle{} }
func (noopHandle) LLMEnd(LLMCall)       {}
func (noopHandle) ToolEnd(ToolCall)     {}
func (noopHandle) Iteration(int)        {}
func (noopHandle) Error(string, error)  {}
func (noopHandle) End(string, error)    {}

// ---- stdout JSON ----

// Stdout writes one JSON object per event to os.Stderr. Useful for local
// debugging; attribute keys follow the OTel GenAI semantic conventions.
type Stdout struct{}

type stdoutHandle struct {
	runID string
	start time.Time
	info  RunInfo
}

func (Stdout) StartRun(info RunInfo) RunHandle {
	h := &stdoutHandle{runID: newUUID(), start: time.Now(), info: info}
	h.emit(map[string]any{
		"event":                "run.start",
		"run.id":               h.runID,
		"session.id":           info.SessionID,
		"gen_ai.system":        info.Provider,
		"gen_ai.request.model": info.Model,
		"tool.protocol":        info.Protocol,
	})
	return h
}

func (h *stdoutHandle) emit(m map[string]any) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}

func (h *stdoutHandle) LLMEnd(c LLMCall) {
	h.emit(map[string]any{
		"event":                      "gen_ai.client.inference",
		"run.id":                     h.runID,
		"gen_ai.system":              c.Provider,
		"gen_ai.request.model":       c.Model,
		"gen_ai.usage.input_tokens":  c.InputTokens,
		"gen_ai.usage.output_tokens": c.OutputTokens,
	})
}

func (h *stdoutHandle) ToolEnd(c ToolCall) {
	h.emit(map[string]any{
		"event":            "gen_ai.execute_tool",
		"run.id":           h.runID,
		"gen_ai.tool.name": c.Name,
		"tool.input":       c.Input,
		"tool.output":      c.Output,
		"error":            c.IsError,
		"duration_ms":      c.Duration.Milliseconds(),
	})
}

func (h *stdoutHandle) Iteration(n int) {
	h.emit(map[string]any{"event": "agent.iteration", "run.id": h.runID, "iteration": n})
}

func (h *stdoutHandle) Error(stage string, err error) {
	h.emit(map[string]any{"event": "error", "run.id": h.runID, "stage": stage, "message": errString(err)})
}

func (h *stdoutHandle) End(output string, err error) {
	h.emit(map[string]any{
		"event":       "run.end",
		"run.id":      h.runID,
		"output":      output,
		"error":       errString(err),
		"duration_ms": time.Since(h.start).Milliseconds(),
	})
}

// ---- LangSmith (REST) ----

// LangSmithConfig configures a LangSmith Observer.
type LangSmithConfig struct {
	APIKey   string // LangSmith API key (x-api-key)
	Project  string // LangSmith project (session_name); defaults to "default"
	Endpoint string // API base; defaults to https://api.smith.langchain.com
	// Client is optional; when nil a client with a 5s timeout is used.
	Client *http.Client
}

// LangSmith posts a completed run to LangSmith on RunHandle.End. Telemetry
// failures are swallowed so they can never break the agent.
type LangSmith struct {
	cfg LangSmithConfig
}

// NewLangSmith builds a LangSmith Observer, applying sensible defaults.
func NewLangSmith(cfg LangSmithConfig) *LangSmith {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.smith.langchain.com"
	}
	if cfg.Project == "" {
		cfg.Project = "default"
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 5 * time.Second}
	}
	return &LangSmith{cfg: cfg}
}

type langsmithHandle struct {
	cfg    LangSmithConfig
	runID  string
	start  time.Time
	info   RunInfo
	mu     sync.Mutex
	tools  []ToolCall
	inTok  int
	outTok int
}

func (l *LangSmith) StartRun(info RunInfo) RunHandle {
	return &langsmithHandle{cfg: l.cfg, runID: newUUID(), start: time.Now(), info: info}
}

func (h *langsmithHandle) LLMEnd(c LLMCall) {
	h.mu.Lock()
	h.inTok += c.InputTokens
	h.outTok += c.OutputTokens
	h.mu.Unlock()
}

func (h *langsmithHandle) ToolEnd(c ToolCall) {
	h.mu.Lock()
	h.tools = append(h.tools, c)
	h.mu.Unlock()
}

func (h *langsmithHandle) Iteration(int)       {}
func (h *langsmithHandle) Error(string, error) {}

func (h *langsmithHandle) End(output string, runErr error) {
	end := time.Now()

	h.mu.Lock()
	toolEvents := make([]map[string]any, 0, len(h.tools))
	for _, t := range h.tools {
		toolEvents = append(toolEvents, map[string]any{
			"name": t.Name, "input": t.Input, "output": t.Output, "error": t.IsError,
		})
	}
	inTok, outTok := h.inTok, h.outTok
	h.mu.Unlock()

	run := map[string]any{
		"id":           h.runID,
		"name":         "darksuitai.agent",
		"run_type":     "chain",
		"start_time":   h.start.UTC().Format(time.RFC3339Nano),
		"end_time":     end.UTC().Format(time.RFC3339Nano),
		"session_name": h.cfg.Project,
		"inputs":       map[string]any{"input": h.info.Input, "session_id": h.info.SessionID},
		"outputs":      map[string]any{"output": output},
		"extra": map[string]any{
			"metadata": map[string]any{
				"gen_ai.system":              h.info.Provider,
				"gen_ai.request.model":       h.info.Model,
				"tool.protocol":              h.info.Protocol,
				"gen_ai.usage.input_tokens":  inTok,
				"gen_ai.usage.output_tokens": outTok,
			},
			"tool_calls": toolEvents,
		},
	}
	if runErr != nil {
		run["error"] = runErr.Error()
	}

	h.post(run)
}

func (h *langsmithHandle) post(run map[string]any) {
	body, err := json.Marshal(run)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, h.cfg.Endpoint+"/runs", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", h.cfg.APIKey)
	resp, err := h.cfg.Client.Do(req)
	if err != nil {
		return
	}
	// Drain and close; telemetry failures are intentionally ignored.
	_ = resp.Body.Close()
}

// ---- helpers ----

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// newUUID returns a random RFC 4122 version 4 UUID string.
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
