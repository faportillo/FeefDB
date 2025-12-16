package core

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewCollection_Invalid(t *testing.T) {
	if _, err := NewCollection("", 2, DistanceMetricDot); err == nil {
		t.Fatalf("expected error for empty name")
	}
	if _, err := NewCollection("c", 0, DistanceMetricDot); err == nil {
		t.Fatalf("expected error for dim <= 0")
	}
}

func TestCollection_Upsert_Search_Flags(t *testing.T) {
	c, err := NewCollection("test", 2, DistanceMetricDot)
	if err != nil {
		t.Fatalf("NewCollection error: %v", err)
	}

	mdA, _ := structpb.NewStruct(map[string]any{"k": "a"})
	mdB, _ := structpb.NewStruct(map[string]any{"k": "b"})
	mdC, _ := structpb.NewStruct(map[string]any{"k": "c"})

	if err := c.Upsert("a", []float32{1, 0}, mdA); err != nil {
		t.Fatalf("Upsert a: %v", err)
	}
	if err := c.Upsert("b", []float32{0, 1}, mdB); err != nil {
		t.Fatalf("Upsert b: %v", err)
	}
	if err := c.Upsert("c", []float32{2, 0}, mdC); err != nil {
		t.Fatalf("Upsert c: %v", err)
	}

	// Query favors x-axis -> expect c then a.
	results, err := c.Search([]float32{1, 0}, 2, true, true)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].ID != "c" || results[1].ID != "a" {
		t.Fatalf("order = [%s, %s], want [c, a]", results[0].ID, results[1].ID)
	}
	if results[0].Vector == nil || results[0].Metadata == nil {
		t.Fatalf("expected vectors and metadata when flags are true")
	}

	// Without vectors/metadata.
	results2, err := c.Search([]float32{1, 0}, 1, false, false)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(results2) != 1 || results2[0].ID != "c" {
		t.Fatalf("unexpected result: %+v", results2)
	}
	if results2[0].Vector != nil || results2[0].Metadata != nil {
		t.Fatalf("expected nil vector/metadata when flags are false")
	}
}

func TestCollection_Delete(t *testing.T) {
	c, _ := NewCollection("test", 2, DistanceMetricCosine)
	_ = c.Upsert("a", []float32{1, 0}, nil)
	if c.Size() != 1 {
		t.Fatalf("size = %d, want 1", c.Size())
	}
	_ = c.Delete("a")
	if c.Size() != 0 {
		t.Fatalf("size = %d, want 0", c.Size())
	}
	res, err := c.Search([]float32{1, 0}, 10, false, false)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("got %d results, want 0", len(res))
	}
}

func TestCollection_DimensionChecks(t *testing.T) {
	c, _ := NewCollection("test", 2, DistanceMetricL2)
	if err := c.Upsert("x", []float32{1, 2, 3}, nil); err == nil {
		t.Fatalf("expected dimension error on Upsert")
	}
	if _, err := c.Search([]float32{1, 2, 3}, 1, false, false); err == nil {
		t.Fatalf("expected dimension error on Search")
	}
}

