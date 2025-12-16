package core

import "testing"

func TestStore_CreateAndGet(t *testing.T) {
	s := NewStore()

	c, err := s.CreateCollection("foo", 2, DistanceMetricDot)
	if err != nil {
		t.Fatalf("CreateCollection error: %v", err)
	}
	if c == nil {
		t.Fatalf("expected collection, got nil")
	}
	if c.Name() != "foo" || c.Dimension() != 2 || c.Metric() != DistanceMetricDot {
		t.Fatalf("unexpected collection fields: name=%s dim=%d metric=%s", c.Name(), c.Dimension(), c.Metric())
	}

	got, err := s.GetCollection("foo")
	if err != nil {
		t.Fatalf("GetCollection error: %v", err)
	}
	if got != c {
		t.Fatalf("expected same pointer for stored collection")
	}
}

func TestStore_CreateDuplicate(t *testing.T) {
	s := NewStore()
	if _, err := s.CreateCollection("dup", 3, DistanceMetricCosine); err != nil {
		t.Fatalf("unexpected error creating first: %v", err)
	}
	if _, err := s.CreateCollection("dup", 3, DistanceMetricCosine); err == nil {
		t.Fatalf("expected duplicate error")
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewStore()
	if _, err := s.GetCollection("missing"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestStore_ListCollections(t *testing.T) {
	s := NewStore()
	_, _ = s.CreateCollection("a", 2, DistanceMetricDot)
	_, _ = s.CreateCollection("b", 2, DistanceMetricL2)

	names := s.ListCollections()
	seen := map[string]bool{}
	for _, n := range names {
		seen[n] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Fatalf("expected names to contain a and b, got %v", names)
	}
}

func TestStore_CreateCollection_InvalidParams(t *testing.T) {
	s := NewStore()
	if _, err := s.CreateCollection("", 2, DistanceMetricDot); err == nil {
		t.Fatalf("expected error for empty name")
	}
	if _, err := s.CreateCollection("bad", 0, DistanceMetricDot); err == nil {
		t.Fatalf("expected error for non-positive dim")
	}
}

