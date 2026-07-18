package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// SummaryStore persists a session's rolling summary and how many leading turns
// it already covers (so compaction never re-summarizes the same turns).
type SummaryStore interface {
	GetSummary(ctx context.Context, sessionID string) (summary string, compactedUpTo int, err error)
	SetSummary(ctx context.Context, sessionID, summary string, compactedUpTo int) error
}

// CompactorConfig tunes when and how aggressively context is compacted.
type CompactorConfig struct {
	// MaxTurns is the number of not-yet-summarized turns tolerated before a
	// compaction pass runs. Defaults to 20.
	MaxTurns int
	// KeepRecent is the number of most-recent turns always kept verbatim
	// (never folded into the summary). Defaults to 6.
	KeepRecent int
}

// Compactor turns a long conversation into a compact context window: a rolling,
// high-fidelity summary of older turns plus the most recent turns verbatim.
// This implements the "compaction" half of context engineering; combined with a
// VectorStore (semantic recall) it keeps the active context small while nothing
// important is lost.
type Compactor struct {
	store      SummaryStore
	summarizer Summarizer
	cfg        CompactorConfig
}

// NewCompactor builds a Compactor, applying default config values.
func NewCompactor(store SummaryStore, summarizer Summarizer, cfg CompactorConfig) *Compactor {
	if cfg.MaxTurns <= 0 {
		cfg.MaxTurns = 20
	}
	if cfg.KeepRecent <= 0 {
		cfg.KeepRecent = 6
	}
	if cfg.KeepRecent > cfg.MaxTurns {
		cfg.KeepRecent = cfg.MaxTurns
	}
	return &Compactor{store: store, summarizer: summarizer, cfg: cfg}
}

// BuildContext returns the compacted context string to inject as chat history.
// allTurns must be the full, chronological (oldest-first) turn list. When the
// backlog of un-summarized turns exceeds MaxTurns, the older ones are folded
// into the rolling summary via the Summarizer and persisted to the SummaryStore.
func (c *Compactor) BuildContext(ctx context.Context, sessionID string, allTurns []Turn) (string, error) {
	summary, upTo, err := c.store.GetSummary(ctx, sessionID)
	if err != nil {
		return "", err
	}
	if upTo < 0 {
		upTo = 0
	}
	if upTo > len(allTurns) {
		upTo = len(allTurns)
	}

	pending := allTurns[upTo:]

	// Compact when the un-summarized backlog is too large.
	if len(pending) > c.cfg.MaxTurns {
		toFold := pending[:len(pending)-c.cfg.KeepRecent]
		newSummary, sErr := c.summarizer.Summarize(ctx, summary, toFold)
		if sErr != nil {
			return "", sErr
		}
		summary = newSummary
		if err := c.store.SetSummary(ctx, sessionID, summary, upTo+len(toFold)); err != nil {
			return "", err
		}
		pending = pending[len(pending)-c.cfg.KeepRecent:]
	}

	return renderContext(summary, pending), nil
}

// renderContext formats the rolling summary and recent turns into a single
// chat-history string.
func renderContext(summary string, recent []Turn) string {
	var b strings.Builder
	if strings.TrimSpace(summary) != "" {
		b.WriteString("Summary of earlier conversation:\n")
		b.WriteString(summary)
		b.WriteString("\n\n")
	}
	if len(recent) > 0 {
		if b.Len() > 0 {
			b.WriteString("Recent conversation:\n")
		}
		for _, t := range recent {
			if t.Human != "" {
				fmt.Fprintf(&b, "Human: %s\n", t.Human)
			}
			if t.AI != "" {
				fmt.Fprintf(&b, "AI: %s\n", t.AI)
			}
		}
	}
	out := strings.TrimRight(b.String(), "\n")
	if out == "" {
		return "[]"
	}
	return out
}

// ---- in-memory summary store ----

type memSummary struct {
	summary string
	upTo    int
}

// InMemorySummaryStore is a process-local SummaryStore for tests and single-node
// use. Production deployments should use the MongoDB-backed store.
type InMemorySummaryStore struct {
	mu   sync.RWMutex
	data map[string]memSummary
}

// NewInMemorySummaryStore returns an empty in-memory summary store.
func NewInMemorySummaryStore() *InMemorySummaryStore {
	return &InMemorySummaryStore{data: make(map[string]memSummary)}
}

// GetSummary returns the stored summary for a session (empty if none).
func (s *InMemorySummaryStore) GetSummary(_ context.Context, sessionID string) (string, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[sessionID]
	if !ok {
		return "", 0, nil
	}
	return v.summary, v.upTo, nil
}

// SetSummary persists the rolling summary and coverage for a session.
func (s *InMemorySummaryStore) SetSummary(_ context.Context, sessionID, summary string, compactedUpTo int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionID] = memSummary{summary: summary, upTo: compactedUpTo}
	return nil
}
