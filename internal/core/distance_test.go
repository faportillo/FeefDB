package core

import "testing"

func TestDotProduct(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{4, 5, 6}
	// 1*4 + 2*5 + 3*6 = 4 + 10 + 18 = 32
	got := DotProduct(a, b)
	if got != 32 {
		t.Fatalf("DotProduct = %v, want 32", got)
	}
}

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float32{1, 2, 3}
	got := CosineSimilarity(a, a)
	// identical vectors should be ~1
	if got < 0.9999 {
		t.Fatalf("CosineSimilarity = %v, want ~1", got)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{0, 0, 0}
	b := []float32{1, 2, 3}
	got := CosineSimilarity(a, b)
	if got != 0 {
		t.Fatalf("CosineSimilarity(zero, b) = %v, want 0", got)
	}
}

func TestNegSquaredL2(t *testing.T) {
	a := []float32{1, 2}
	b := []float32{4, 6}
	// squared distance: (1-4)^2 + (2-6)^2 = 9 + 16 = 25 => neg is -25
	got := NegSquaredL2(a, b)
	if got != -25 {
		t.Fatalf("NegSquaredL2 = %v, want -25", got)
	}
}

func TestParseDistanceMetric(t *testing.T) {
	m, err := ParseDistanceMetric("  COSINE ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != DistanceMetricCosine {
		t.Fatalf("got %q, want %q", m, DistanceMetricCosine)
	}
}
