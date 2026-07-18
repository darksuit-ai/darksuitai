package memory

import (
	"context"
	"testing"
)

func TestCosine(t *testing.T) {
	cases := []struct {
		name string
		a, b []float32
		want float64
	}{
		{"identical", []float32{1, 0, 0}, []float32{1, 0, 0}, 1},
		{"orthogonal", []float32{1, 0}, []float32{0, 1}, 0},
		{"opposite", []float32{1, 0}, []float32{-1, 0}, -1},
		{"mismatched len", []float32{1, 0}, []float32{1, 0, 0}, 0},
		{"zero vector", []float32{0, 0}, []float32{1, 1}, 0},
	}
	for _, c := range cases {
		if got := Cosine(c.a, c.b); got < c.want-1e-6 || got > c.want+1e-6 {
			t.Errorf("%s: Cosine=%v want %v", c.name, got, c.want)
		}
	}
}

func TestInMemoryVectorStore_SearchRanking(t *testing.T) {
	ctx := context.Background()
	s := NewInMemoryVectorStore()
	_ = s.Add(ctx, "apple", "about apples", []float32{1, 0, 0}, nil)
	_ = s.Add(ctx, "banana", "about bananas", []float32{0, 1, 0}, nil)
	_ = s.Add(ctx, "applish", "also applish", []float32{0.9, 0.1, 0}, nil)

	hits, err := s.Search(ctx, []float32{1, 0, 0}, 2)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("want 2 hits, got %d", len(hits))
	}
	if hits[0].ID != "apple" {
		t.Errorf("want top hit 'apple', got %q", hits[0].ID)
	}
	if hits[1].ID != "applish" {
		t.Errorf("want second hit 'applish', got %q", hits[1].ID)
	}
	if hits[0].Score < hits[1].Score {
		t.Errorf("results not sorted by score desc: %v < %v", hits[0].Score, hits[1].Score)
	}
}

func TestInMemoryVectorStore_ReplaceByID(t *testing.T) {
	ctx := context.Background()
	s := NewInMemoryVectorStore()
	_ = s.Add(ctx, "x", "v1", []float32{1, 0}, nil)
	_ = s.Add(ctx, "x", "v2", []float32{0, 1}, nil)
	hits, _ := s.Search(ctx, []float32{0, 1}, 5)
	if len(hits) != 1 {
		t.Fatalf("want 1 entry after replace, got %d", len(hits))
	}
	if hits[0].Text != "v2" {
		t.Errorf("want replaced text 'v2', got %q", hits[0].Text)
	}
}

func TestInMemoryVectorStore_EmptyVector(t *testing.T) {
	if err := NewInMemoryVectorStore().Add(context.Background(), "id", "t", nil, nil); err == nil {
		t.Error("expected error adding empty vector")
	}
}
