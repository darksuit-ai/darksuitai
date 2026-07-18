// Package memory provides context and memory engineering primitives (Phase 4):
// conversation compaction (rolling summarization), and semantic retrieval over
// a vector store. The core types here are provider-agnostic interfaces; concrete
// adapters (MongoDB stores, an Anthropic summarizer, an HTTP embedder) live in
// sibling packages so this package stays dependency-free and unit-testable.
package memory

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
)

// Turn is a single conversational exchange.
type Turn struct {
	Human string
	AI    string
}

// Summarizer condenses prior context into a compact, high-fidelity summary.
// priorSummary may be empty on the first compaction; turns are the (older)
// turns being folded into the summary.
type Summarizer interface {
	Summarize(ctx context.Context, priorSummary string, turns []Turn) (string, error)
}

// Embedder converts text into a vector embedding for semantic retrieval.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Hit is a semantic-search result.
type Hit struct {
	ID    string
	Text  string
	Score float64
	Meta  map[string]any
}

// VectorStore persists text embeddings and retrieves the nearest neighbours.
type VectorStore interface {
	Add(ctx context.Context, id, text string, vector []float32, meta map[string]any) error
	Search(ctx context.Context, vector []float32, k int) ([]Hit, error)
}

// ---- cosine similarity ----

// Cosine returns the cosine similarity of two equal-length vectors in [-1, 1].
// It returns 0 for mismatched lengths or zero-magnitude vectors.
func Cosine(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// ---- in-memory vector store ----

type memVector struct {
	id   string
	text string
	vec  []float32
	meta map[string]any
}

// InMemoryVectorStore is a simple cosine-ranked VectorStore. It is safe for
// concurrent use and is intended for tests and small/local deployments; use the
// MongoDB Atlas vector store for production-scale retrieval.
type InMemoryVectorStore struct {
	mu      sync.RWMutex
	vectors []memVector
}

// NewInMemoryVectorStore returns an empty in-memory vector store.
func NewInMemoryVectorStore() *InMemoryVectorStore { return &InMemoryVectorStore{} }

// Add stores (or replaces, by id) an embedded text entry.
func (s *InMemoryVectorStore) Add(_ context.Context, id, text string, vector []float32, meta map[string]any) error {
	if len(vector) == 0 {
		return errors.New("memory: empty vector")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.vectors {
		if s.vectors[i].id == id {
			s.vectors[i] = memVector{id: id, text: text, vec: vector, meta: meta}
			return nil
		}
	}
	s.vectors = append(s.vectors, memVector{id: id, text: text, vec: vector, meta: meta})
	return nil
}

// Search returns the k most cosine-similar entries, highest score first.
func (s *InMemoryVectorStore) Search(_ context.Context, vector []float32, k int) ([]Hit, error) {
	if k <= 0 {
		k = 5
	}
	s.mu.RLock()
	hits := make([]Hit, 0, len(s.vectors))
	for _, v := range s.vectors {
		hits = append(hits, Hit{ID: v.id, Text: v.text, Score: Cosine(vector, v.vec), Meta: v.meta})
	}
	s.mu.RUnlock()

	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if len(hits) > k {
		hits = hits[:k]
	}
	return hits, nil
}
