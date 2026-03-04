package pbf

import (
	"bytes"
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

func TestNewEncoderFailsOnInvalidStorePath(t *testing.T) {
	invalidStore := filepath.Join(t.TempDir(), "missing", "store")

	enc, err := NewEncoder(&bytes.Buffer{}, WithStorePath(invalidStore))
	if err == nil {
		t.Fatal("expected NewEncoder to fail for invalid store path")
	}
	if enc != nil {
		t.Fatal("expected nil encoder when setup fails")
	}
	if !errors.Is(err, ErrCreateTempFile) {
		t.Fatalf("expected ErrCreateTempFile, got: %v", err)
	}
}

func TestNewEncoderSetsSpecRequiredFeatures(t *testing.T) {
	const (
		requiredFeatureOSMSchema             = "OsmSchema-V0.6"
		requiredFeatureDenseNodes            = "DenseNodes"
		requiredFeatureHistoricalInformation = "HistoricalInformation"
	)

	enc, err := NewEncoder(&bytes.Buffer{})
	if err != nil {
		t.Fatalf("create encoder: %v", err)
	}
	defer enc.Close()

	required := enc.Header.RequiredFeatures

	if !slices.Contains(required, requiredFeatureOSMSchema) {
		t.Fatalf("missing required feature %q in %#v", requiredFeatureOSMSchema, required)
	}
	if !slices.Contains(required, requiredFeatureDenseNodes) {
		t.Fatalf("missing required feature %q in %#v", requiredFeatureDenseNodes, required)
	}
	if !slices.Contains(required, requiredFeatureHistoricalInformation) {
		t.Fatalf("missing required feature %q in %#v", requiredFeatureHistoricalInformation, required)
	}
}
