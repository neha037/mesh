package ollama

import (
	"math"
	"testing"
)

func TestValidateEmbedding_NaN(t *testing.T) {
	v := make([]float32, 10)
	v[0] = float32(math.NaN())
	err := validateEmbedding(v)
	if err == nil {
		t.Fatal("expected error for NaN embedding")
	}
	want := "embedding contains NaN at index 0"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestValidateEmbedding_Inf(t *testing.T) {
	v := make([]float32, 20)
	v[10] = float32(math.Inf(1))
	err := validateEmbedding(v)
	if err == nil {
		t.Fatal("expected error for Inf embedding")
	}
	want := "embedding contains Inf at index 10"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestValidateEmbedding_AllZeros(t *testing.T) {
	v := make([]float32, 10)
	err := validateEmbedding(v)
	if err == nil {
		t.Fatal("expected error for all-zeros embedding")
	}
	want := "embedding is all zeros"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestValidateEmbedding_Valid(t *testing.T) {
	v := []float32{0.1, 0.2, 0.3, 0.4}
	if err := validateEmbedding(v); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNormalizeEmbedding(t *testing.T) {
	v := []float32{3.0, 4.0, 0.0}
	normalizeEmbedding(v)
	// L2 norm should be approximately 1.0
	var norm float64
	for _, f := range v {
		norm += float64(f) * float64(f)
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 1e-6 {
		t.Errorf("L2 norm = %f, want 1.0", norm)
	}
}
