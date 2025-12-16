package core

import (
	"fmt"
	"math"
	"strings"
)

// DistanceFunc returns a score where HIGHER is better.
// - Cosine: higher cosine similarity = better
// - Dot: higher dot product = better
// - L2: lower distance is better, so we return NEGATIVE squared L2 (higher = better)
type DistanceFunc func(a, b []float32) float32
type DistanceMetric string

const (
	DistanceMetricCosine DistanceMetric = "cosine"
	DistanceMetricDot    DistanceMetric = "dot"
	DistanceMetricL2     DistanceMetric = "l2"
)

// ParseDistanceMetric normalizes/validates a user-provided metric name.
// This is useful in your CreateCollection RPC where distance comes in as a string.
func ParseDistanceMetric(s string) (DistanceMetric, error) {
	switch strings.ToLower(strings.TrimSpace(s)){
		case string(DistanceMetricCosine):
			return DistanceMetricCosine, nil
		case string(DistanceMetricDot):
			return DistanceMetricDot, nil
		case string(DistanceMetricL2):
			return DistanceMetricL2, nil
		default:
			return "", fmt.Errorf("unknown distance metric: %q (expected: cosine, dot, l2)", s)
	}
}

// DistanceFuncForMetric returns the appropriate DistanceFunc for a given metric.
// This is useful for your index/collection to use the correct distance calculation.
func DistanceFuncForMetric(metric DistanceMetric) (DistanceFunc, error) {
	switch metric{
		case DistanceMetricCosine:
			return CosineSimilarity, nil
		case DistanceMetricDot:
			return DotProduct, nil
		case DistanceMetricL2:
			return NegSquaredL2, nil
		default:
			return nil, fmt.Errorf("unsupported distance metric: %q", metric)
	}
}

// DotProduct computes sum(a[i] * b[i]).
// Higher is better.
func DotProduct(a, b []float32) float32 {
	// We do NOT check length equality here for speed.
	// Caller (index/collection) should enforce consistent dimension.
	var dot float32
	for i := range a{
		dot += a[i] * b[i]
	}
	return dot
}

// CosineSimilarity computes dot(a,b)/(||a||*||b||).
// Range: [-1, 1]. Higher is better.
// Edge case: if either vector is all zeros, returns 0.
func CosineSimilarity(a, b []float32) float32 {
	var dot, na, nb float32
	for i := range a{
		ai := a[i]
		bi := b[i]
		dot += ai * bi
		na += ai * ai
		nb += bi * bi
	}

	// Handle 0 vectors gracefully.
	if na == 0 || nb == 0{
		return 0
	}

	denom := float32(math.Sqrt(float64(na) * float64(nb)))
	if denom == 0{
		return 0
	}
	return dot / denom
}

// NegSquaredL2 computes -sum((a[i]-b[i])^2).
// We use squared L2 (no sqrt) because sqrt is monotonic and unnecessary.
// Higher is better because we negate it.
func NegSquaredL2(a, b []float32) float32 {
	var sum float32
	for i := range a{
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return -sum
}