package core

import (
	"container/heap"
	"fmt"
	"sync"
)

// BruteForceIndex is the simplest possible index: scan every vector at query time.
// Great for correctness + learning. Slow at large N.
//
// Thread safety:
// - Uses RWMutex to allow concurrent searches (read lock) while upserts acquire write lock.
// - Maps in Go are not safe for concurrent access without locks.
type BruteForceIndex struct {
	dim int
	metric DistanceMetric
	dist DistanceFunc

	mu sync.RWMutex
	vectors map[string][]float32
}

func NewBruteForceIndex(dim int, metric DistanceMetric) (*BruteForceIndex, error) {
	if dim <= 0 {
		return nil, fmt.Errorf("dimension must be > 0, got %d", dim)
	}

	df, err := DistanceFuncForMetric(metric)
	if err != nil {
		return nil, err
	}

	return &BruteForceIndex{
		dim: dim,
		metric: metric,
		dist: df,
		vectors: make(map[string][]float32),
	}, nil
	
}

func (b *BruteForceIndex) Dimension() int {
	return b.dim
}

func (b *BruteForceIndex) Metric() DistanceMetric {
	return b.metric
}

func (bf *BruteForceIndex) Size() int {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	return len(bf.vectors)
}

func (bf *BruteForceIndex) Add(id string, vector []float32) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if err := CheckDim(bf.dim, vector); err != nil {
		return err
	}

	copiedVec := make([]float32, bf.dim)
	copy(copiedVec, vector)

	bf.mu.Lock()
	defer bf.mu.Unlock()
	bf.vectors[id] = copiedVec
	return nil
}

// Delete removes a vector by id. Not an error if it doesn't exist.
func (bf* BruteForceIndex) Delete(id string) error {
	if id == "" {
		return nil
	}

	bf.mu.Lock()
	defer bf.mu.Unlock()

	delete(bf.vectors, id)
	return nil
}

// Search returns up to k nearest neighbors (best-first).
//
// Implementation detail:
// We keep a MIN-HEAP of size k. The smallest score is at the root.
// For each candidate:
// - if heap not full: push
// - else if candidate.score > min.score: pop min, push candidate
//
// Complexity:
// - Distance compute: O(N * dim)
// - Heap ops: O(N log k)
func (bf* BruteForceIndex) Search(query []float32, k int) ([]SearchResult, error){
	if err := CheckK(k); err != nil {
		return nil, err
	}
	if err := CheckDim(bf.dim, query); err != nil {
		return nil, err
	}

	bf.mu.RLock()
	defer bf.mu.RUnlock()

	if len(bf.vectors) == 0 {
		return []SearchResult{}, nil
	}

	h := &minKHeap{}
	heap.Init(h)

	for id, vec := range bf.vectors {
		score := bf.dist(query, vec)
		item := SearchResult{ID: id, Score: score}

		if h.Len() < k {
			heap.Push(h, item)
			continue
		}

		if score > (*h)[0].Score {
			heap.Pop(h)
			heap.Push(h, item)
		}
	}

	// Extract heap into a slice sorted best-first.
	// Heap pops smallest first, so we fill from end to start.
	out := make([]SearchResult, h.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(h).(SearchResult)
	}
	return out, nil
}

type minKHeap []SearchResult
func (h minKHeap) Len() int {return len(h)}
func (h minKHeap) Less(i, j int) bool {
	return h[i].Score < h[j].Score
}
func (h minKHeap) Swap(i, j int) {h[i], h[j] = h[j], h[i]}
func (h *minKHeap) Push(x any) {
	*h = append(*h, x.(SearchResult))
}
func (h *minKHeap) Pop() any{
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}