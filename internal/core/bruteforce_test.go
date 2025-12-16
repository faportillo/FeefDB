package core

import "testing"

func TestBruteForceIndex_AddAndSearch(t *testing.T) {
	idx, err := NewBruteForceIndex(2, DistanceMetricDot)
	if err != nil {
		t.Fatalf("NewBruteForceIndex error: %v", err)
	}

	_ = idx.Add("a", []float32{1, 0})
	_ = idx.Add("b", []float32{0, 1})
	_ = idx.Add("c", []float32{2, 0})

	// Query points more in x direction => c should be best, then a, then b.
	res, err := idx.Search([]float32{1, 0}, 2)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d results, want 2", len(res))
	}
	if res[0].ID != "c" {
		t.Fatalf("best result = %q, want %q", res[0].ID, "c")
	}
	if res[1].ID != "a" {
		t.Fatalf("second result = %q, want %q", res[1].ID, "a")
	}
}

func TestBruteForceIndex_DimensionMismatch(t *testing.T) {
	idx, _ := NewBruteForceIndex(2, DistanceMetricCosine)

	if err := idx.Add("a", []float32{1, 2, 3}); err == nil {
		t.Fatalf("expected dimension error")
	}

	if _, err := idx.Search([]float32{1, 2, 3}, 5); err == nil {
		t.Fatalf("expected dimension error")
	}
}

func TestBruteForceIndex_Delete(t *testing.T) {
	idx, _ := NewBruteForceIndex(2, DistanceMetricCosine)

	_ = idx.Add("a", []float32{1, 0})
	_ = idx.Delete("a")
	res, err := idx.Search([]float32{1, 0}, 10)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("got %d results, want 0", len(res))
	}
}
