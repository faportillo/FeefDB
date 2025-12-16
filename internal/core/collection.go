package core

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"
)

// Point is the stored record for an ID.
// The Index only needs Vector, but Collection stores metadata too.
type Point struct {
	ID	string
	Vector []float32
	Metadata *structpb.Struct
}

type ScoredPoint struct {
	ID       string
	Score    float32
	Vector   []float32
	Metadata *structpb.Struct
}

// Collection holds:
// - an Index (nearest neighbor engine)
// - a point store (authoritative vectors + metadata)
// - collection-level config (dimension + metric)
type Collection struct {
	name   string
	dim    int
	metric DistanceMetric
	index  Index

	mu     sync.RWMutex
	points map[string]*Point
}

func NewCollection(name string, dim int, metric DistanceMetric) (*Collection, error) {
	if name == "" {
		return nil, fmt.Errorf("collection name must not be empty")
	}

	if dim <= 0 {
		return nil, fmt.Errorf("dim must be > 0, got %d", dim)
	}

	// Swap for other Index Algos later (IVF, HNSW)
	idx, err := NewBruteForceIndex(dim, metric)
	if err != nil {
		return nil, err
	}

	return &Collection{
		name:   name,
		dim:    dim,
		metric: metric,
		index:  idx,
		points: make(map[string]*Point),
	}, nil
}

func (c *Collection) Name() string {return c.name}
func (c *Collection) Dimension() int {return c.dim}
func (c *Collection) Metric() DistanceMetric {return c.metric}
func (c *Collection) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.points)
}

// Upsert inserts or overwrites a point.
//
// Design note: we update the point store and the index.
// Point store is the source of truth for metadata.
// Index is the source of truth for nearest-neighbor scoring.
func (c *Collection) Upsert(id string, vec []float32, md *structpb.Struct) error {
	if id == "" {
		return fmt.Errorf("id must not be empty")
	}

	if err := CheckDim(c.dim, vec); err != nil {
		return err
	} 

	// Copy vector so caller can’t mutate internal state.
	copied := make([]float32, c.dim)
	copy(copied, vec)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.points[id] = &Point{
		ID: id,
		Vector: copied,
		Metadata: md,
	}

	return c.index.Add(id, copied)
}

func (c *Collection) Delete(id string) error {
	if id == "" {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.points, id)

	return c.index.Delete(id)
}

// Search runs nearest-neighbor search and optionally includes vectors/metadata.
//
// includeVectors/includeMetadata correspond directly proto flags.
// TODO: filter is not implemented yet here; can add it next.
func (c *Collection) Search(query []float32, k int, includeVectors, includeMetadata bool) ([]ScoredPoint, error) {
	if err := CheckDim(c.dim, query); err != nil {
		return nil, err
	}
	if err := CheckK(k); err != nil {
		return nil, err
	}

	// Step 1: ask the index for nearest IDs + scores.
	raw, err := c.index.Search(query, k)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return []ScoredPoint{}, nil
	}

	// Step 2: attach stored data (vector/metadata) from the point store.
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]ScoredPoint, 0, len(raw))
	for _, r := range raw {
		p := c.points[r.ID]
		if p == nil {
			// Should be rare (if store/index get out of sync).
			// For now, just skip.
			continue
		}

		sp := ScoredPoint{
			ID:    r.ID,
			Score: r.Score,
		}
		if includeVectors {
			sp.Vector = p.Vector
		}
		if includeMetadata {
			sp.Metadata = p.Metadata
		}
		out = append(out, sp)
	}

	return out, nil
}