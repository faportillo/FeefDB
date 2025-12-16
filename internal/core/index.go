package core

import "fmt"

type Index interface {
	Dimension() int
	Metric() DistanceMetric

	Add(id string, vector []float32) error
	Search(query []float32, topK int) ([]SearchResult, error)
	Delete(id string) error
	// Size returns the number of stored vectors.
	Size() int
}

// SearchResult is a scored neighbor returned by an Index.
// Score must be comparable such that higher means better.
type SearchResult struct{
	ID string
	Score float32
}

// Common errors you can reuse across implementations.
var (
	ErrInvalidK          = fmt.Errorf("k must be > 0")
	ErrDimensionMismatch = fmt.Errorf("dimension mismatch")
)

// CheckDim is a small helper used by implementations to validate vector length.
// Returning a shared error makes it easy for the service layer to map it to gRPC codes.
func CheckDim(expected int, vector []float32) error {
	if expected != len(vector){
		return ErrDimensionMismatch
	}
	return nil
}

// CheckK validates k for searches.
func CheckK(k int) error {
	if k <= 0 {
		return ErrInvalidK
	}
	return nil
}