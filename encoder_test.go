package pbf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"m4o.io/pbf/v2/model"
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

func TestRoundTripPreservesDenseNodeVisibility(t *testing.T) {
	t.Parallel()

	nodes := []model.Entity{
		&model.Node{
			ID:   1001,
			Lat:  42.1234567,
			Lon:  -71.9876543,
			Tags: map[string]string{"name": "hidden"},
			Info: &model.Info{
				Version:   3,
				UID:       101,
				Timestamp: time.Unix(1_700_000_000, 0).UTC(),
				Changeset: 55,
				User:      "alice",
				Visible:   false,
			},
		},
		&model.Node{
			ID:   1002,
			Lat:  42.2234567,
			Lon:  -71.8876543,
			Tags: map[string]string{"name": "shown"},
			Info: &model.Info{
				Version:   4,
				UID:       102,
				Timestamp: time.Unix(1_700_000_500, 0).UTC(),
				Changeset: 56,
				User:      "bob",
				Visible:   true,
			},
		},
	}

	var encoded bytes.Buffer

	enc, err := NewEncoder(&encoded)
	if err != nil {
		t.Fatalf("create encoder: %v", err)
	}

	if err := enc.EncodeBatch(nodes); err != nil {
		t.Fatalf("encode nodes: %v", err)
	}
	enc.Close()

	dec, err := NewDecoder(context.Background(), bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatalf("create decoder: %v", err)
	}
	defer dec.Close()

	gotVisibility := map[model.ID]bool{}

	for {
		batch, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("decode batch: %v", err)
		}

		for _, entity := range batch {
			node, ok := entity.(*model.Node)
			if !ok {
				t.Fatalf("expected *model.Node, got %T", entity)
			}

			gotVisibility[node.ID] = node.Info.Visible
		}
	}

	if got, ok := gotVisibility[1001]; !ok {
		t.Fatalf("missing node 1001 from decoded output: %#v", gotVisibility)
	} else if got {
		t.Fatalf("node 1001 visibility mismatch: got %t want %t", got, false)
	}
	if got, ok := gotVisibility[1002]; !ok {
		t.Fatalf("missing node 1002 from decoded output: %#v", gotVisibility)
	} else if !got {
		t.Fatalf("node 1002 visibility mismatch: got %t want %t", got, true)
	}
}
