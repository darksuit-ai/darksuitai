package memory

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// fakeSummarizer records how many turns it was asked to fold and returns a
// deterministic summary so tests can assert compaction behaviour.
type fakeSummarizer struct {
	calls  int
	folded int
}

func (f *fakeSummarizer) Summarize(_ context.Context, prior string, turns []Turn) (string, error) {
	f.calls++
	f.folded += len(turns)
	return fmt.Sprintf("summary(prior=%q,+%d turns)", prior, len(turns)), nil
}

func turns(n int) []Turn {
	out := make([]Turn, n)
	for i := 0; i < n; i++ {
		out[i] = Turn{Human: fmt.Sprintf("q%d", i), AI: fmt.Sprintf("a%d", i)}
	}
	return out
}

func TestCompactor_NoCompactionUnderThreshold(t *testing.T) {
	ctx := context.Background()
	sum := &fakeSummarizer{}
	c := NewCompactor(NewInMemorySummaryStore(), sum, CompactorConfig{MaxTurns: 10, KeepRecent: 4})

	out, err := c.BuildContext(ctx, "s1", turns(5))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if sum.calls != 0 {
		t.Errorf("expected no summarization under threshold, got %d calls", sum.calls)
	}
	if !strings.Contains(out, "Human: q0") || !strings.Contains(out, "AI: a4") {
		t.Errorf("expected verbatim recent turns, got:\n%s", out)
	}
}

func TestCompactor_CompactsOverThreshold(t *testing.T) {
	ctx := context.Background()
	sum := &fakeSummarizer{}
	store := NewInMemorySummaryStore()
	c := NewCompactor(store, sum, CompactorConfig{MaxTurns: 10, KeepRecent: 4})

	// 15 pending turns > MaxTurns(10): fold 15-4=11, keep 4 recent.
	out, err := c.BuildContext(ctx, "s1", turns(15))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if sum.calls != 1 {
		t.Fatalf("expected 1 summarization, got %d", sum.calls)
	}
	if sum.folded != 11 {
		t.Errorf("expected 11 turns folded, got %d", sum.folded)
	}
	if !strings.Contains(out, "Summary of earlier conversation:") {
		t.Errorf("expected summary section, got:\n%s", out)
	}
	// Recent 4 turns kept verbatim (q11..q14); older ones summarized away.
	if !strings.Contains(out, "Human: q14") {
		t.Errorf("expected recent turn q14 verbatim, got:\n%s", out)
	}
	if strings.Contains(out, "Human: q0") {
		t.Errorf("old turn q0 should have been compacted away, got:\n%s", out)
	}

	// Coverage should be persisted.
	_, upTo, _ := store.GetSummary(ctx, "s1")
	if upTo != 11 {
		t.Errorf("expected compactedUpTo=11, got %d", upTo)
	}
}

func TestCompactor_DoesNotResummarizeCoveredTurns(t *testing.T) {
	ctx := context.Background()
	sum := &fakeSummarizer{}
	c := NewCompactor(NewInMemorySummaryStore(), sum, CompactorConfig{MaxTurns: 10, KeepRecent: 4})

	all := turns(15)
	if _, err := c.BuildContext(ctx, "s1", all); err != nil {
		t.Fatalf("first: %v", err)
	}
	foldedAfterFirst := sum.folded

	// Add 2 more turns; pending backlog (4 kept + 2 new = 6) is under MaxTurns,
	// so no new compaction should occur.
	if _, err := c.BuildContext(ctx, "s1", turns(17)); err != nil {
		t.Fatalf("second: %v", err)
	}
	if sum.folded != foldedAfterFirst {
		t.Errorf("covered turns were re-summarized: folded went %d -> %d", foldedAfterFirst, sum.folded)
	}
}

func TestCompactor_EmptyIsSentinel(t *testing.T) {
	out, err := NewCompactor(NewInMemorySummaryStore(), &fakeSummarizer{}, CompactorConfig{}).
		BuildContext(context.Background(), "s1", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out != "[]" {
		t.Errorf("expected sentinel [] for empty history, got %q", out)
	}
}
