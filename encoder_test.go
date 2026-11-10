// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/encoder"
	"m4o.io/pbf/v2/internal/pb"
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
	if slices.Contains(required, requiredFeatureHistoricalInformation) {
		t.Fatalf("unexpected required feature %q in %#v", requiredFeatureHistoricalInformation, required)
	}
}

func TestNewEncoderFeatureOptionsAppendAndDeduplicate(t *testing.T) {
	t.Parallel()

	defaults := []string{"OsmSchema-V0.6", "DenseNodes"}

	testCases := []struct {
		name         string
		opts         []EncoderOption
		wantRequired []string
		wantOptional []string
	}{
		{
			name:         "required keeps defaults",
			opts:         []EncoderOption{WithRequiredFeatures("Acme-Custom")},
			wantRequired: append(slices.Clone(defaults), "Acme-Custom"),
		},
		{
			name:         "required removes duplicates of defaults",
			opts:         []EncoderOption{WithRequiredFeatures("DenseNodes", "Acme-Custom", "OsmSchema-V0.6")},
			wantRequired: append(slices.Clone(defaults), "Acme-Custom"),
		},
		{
			name: "required accumulates across calls",
			opts: []EncoderOption{
				WithRequiredFeatures("Acme-A", "Acme-B"),
				WithRequiredFeatures("Acme-B", "Acme-C"),
			},
			wantRequired: append(slices.Clone(defaults), "Acme-A", "Acme-B", "Acme-C"),
		},
		{
			name: "optional accumulates and removes duplicates",
			opts: []EncoderOption{
				WithOptionalFeatures("Has_Metadata", "Sort.Type_then_ID"),
				WithOptionalFeatures("Sort.Type_then_ID"),
			},
			wantRequired: defaults,
			wantOptional: []string{"Has_Metadata", "Sort.Type_then_ID"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			enc, err := NewEncoder(&bytes.Buffer{}, tc.opts...)
			if err != nil {
				t.Fatalf("create encoder: %v", err)
			}
			defer enc.Close()

			if got := enc.Header.RequiredFeatures; !slices.Equal(got, tc.wantRequired) {
				t.Errorf("required features = %q, want %q", got, tc.wantRequired)
			}
			if got := enc.Header.OptionalFeatures; !slices.Equal(got, tc.wantOptional) {
				t.Errorf("optional features = %q, want %q", got, tc.wantOptional)
			}
		})
	}
}

// TestRoundTripDropsDenseNodeVisibility checks that each node decodes with
// Visible set to true, also the node that the test encodes with Visible set to
// false. The encoder does not write the visible flag, and the decoder sets
// Visible to true when the flag is not there.
func TestRoundTripDropsDenseNodeVisibility(t *testing.T) {
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
	} else if !got {
		t.Fatalf("node 1001 visibility mismatch: got %t want %t", got, true)
	}
	if got, ok := gotVisibility[1002]; !ok {
		t.Fatalf("missing node 1002 from decoded output: %#v", gotVisibility)
	} else if !got {
		t.Fatalf("node 1002 visibility mismatch: got %t want %t", got, true)
	}
}

func TestRoundTripPreservesOsmosisReplicationTimestamp(t *testing.T) {
	t.Parallel()

	want := time.Date(2024, time.October, 28, 21, 21, 30, 0, time.UTC)

	nodes := []model.Entity{
		&model.Node{ID: 1, Info: &model.Info{Visible: true}},
	}

	var encoded bytes.Buffer

	enc, err := NewEncoder(&encoded, WithOsmosisReplicationTimestamp(want))
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

	if got := dec.Header.OsmosisReplicationTimestamp; !got.Equal(want) {
		t.Fatalf("replication timestamp = %v, want %v", got, want)
	}
}

func TestEncoderKeepsCallerOrder(t *testing.T) {
	t.Parallel()

	info := func() *model.Info {
		return &model.Info{Timestamp: time.Unix(1_700_000_000, 0).UTC(), Visible: true}
	}

	want := []model.Entity{
		&model.Relation{ID: 1, Info: info()},
		&model.Node{ID: 1, Info: info()},
		&model.Way{ID: 1, Info: info()},
		&model.Node{ID: 2, Info: info()},
		&model.Node{ID: 3, Info: info()},
		&model.Relation{ID: 2, Info: info()},
		&model.Way{ID: 2, Info: info()},
	}

	encode := func() []byte {
		var encoded bytes.Buffer

		enc, err := NewEncoder(&encoded)
		require.NoError(t, err)

		for _, entity := range want {
			require.NoError(t, enc.Encode(entity))
		}

		enc.Close()

		return encoded.Bytes()
	}

	encoded := encode()
	assert.Equal(t, encoded, encode(), "two encodes of the same entities give different bytes")

	dec, err := NewDecoder(context.Background(), bytes.NewReader(encoded))
	require.NoError(t, err)

	defer dec.Close()

	var got []model.Entity

	for {
		batch, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		got = append(got, batch...)
	}

	require.Len(t, got, len(want))

	for i := range want {
		assert.IsType(t, want[i], got[i], "entity %d", i)
		assert.Equal(t, want[i].GetID(), got[i].GetID(), "entity %d", i)
	}
}

func TestRoundTripWithEachCompression(t *testing.T) {
	t.Parallel()

	info := func() *model.Info {
		return &model.Info{
			Version:   2,
			UID:       7,
			Timestamp: time.Unix(1_700_000_000, 0).UTC(),
			Changeset: 9,
			User:      "carol",
			Visible:   true,
		}
	}

	want := []model.Entity{
		&model.Node{ID: 1, Tags: map[string]string{"name": "a"}, Info: info(), Lat: 42.5, Lon: -71.25},
		&model.Node{ID: 2, Tags: map[string]string{}, Info: info(), Lat: 42.75, Lon: -71.5},
		&model.Way{ID: 10, Tags: map[string]string{"highway": "path"}, Info: info(), NodeIDs: []model.ID{1, 2}},
		&model.Relation{
			ID:   20,
			Tags: map[string]string{"type": "route"},
			Info: info(),
			Members: []model.Member{
				{ID: 10, Type: model.WAY, Role: "forward"},
				{ID: 1, Type: model.NODE, Role: "stop"},
			},
		},
	}

	tests := []struct {
		compression encoder.BlobCompression
		wantData    any
	}{
		{encoder.RAW, &pb.Blob_Raw{}},
		{encoder.ZLIB, &pb.Blob_ZlibData{}},
		{encoder.LZMA, &pb.Blob_LzmaData{}},
		{encoder.LZ4, &pb.Blob_Lz4Data{}},
		{encoder.ZSTD, &pb.Blob_ZstdData{}},
	}

	for _, tt := range tests {
		t.Run(tt.compression.String(), func(t *testing.T) {
			t.Parallel()

			var encoded bytes.Buffer

			compression := WithCompression(tt.compression)

			enc, err := NewEncoder(&encoded, compression)
			require.NoError(t, err)

			require.NoError(t, enc.EncodeBatch(want))
			enc.Close()

			blobs := readBlobs(t, encoded.Bytes())
			require.NotEmpty(t, blobs)

			for i, blob := range blobs {
				assert.IsType(t, tt.wantData, blob.GetData(), "blob %d", i)
			}

			got := decodeAll(t, encoded.Bytes())
			assert.Equal(t, want, got)
		})
	}
}

func TestRoundTripPreservesHeaderOptions(t *testing.T) {
	t.Parallel()

	const (
		wantProgram        = "pbf-test/1.0"
		wantSource         = "https://example.com/source"
		wantSequenceNumber = int64(4221)
		wantBaseURL        = "https://example.com/replication"
	)

	opts := []EncoderOption{
		WithWritingProgram(wantProgram),
		WithSource(wantSource),
		WithOsmosisReplicationSequenceNumber(wantSequenceNumber),
		WithOsmosisReplicationBaseURL(wantBaseURL),
	}

	var encoded bytes.Buffer

	enc, err := NewEncoder(&encoded, opts...)
	require.NoError(t, err)

	node := &model.Node{ID: 1, Info: &model.Info{Visible: true}}
	require.NoError(t, enc.Encode(node))
	enc.Close()

	r := bytes.NewReader(encoded.Bytes())

	dec, err := NewDecoder(context.Background(), r)
	require.NoError(t, err)

	defer dec.Close()

	assert.Equal(t, wantProgram, dec.Header.WritingProgram)
	assert.Equal(t, wantSource, dec.Header.Source)
	assert.Equal(t, wantSequenceNumber, dec.Header.OsmosisReplicationSequenceNumber)
	assert.Equal(t, wantBaseURL, dec.Header.OsmosisReplicationBaseURL)
}

// readBlobs reads each blob of a PBF stream. readBlobs does not uncompress
// the data of the blobs.
func readBlobs(t *testing.T, encoded []byte) []*pb.Blob {
	t.Helper()

	r := bytes.NewReader(encoded)

	var blobs []*pb.Blob

	for r.Len() > 0 {
		var size uint32
		require.NoError(t, binary.Read(r, binary.BigEndian, &size))

		headerBytes := make([]byte, size)
		_, err := io.ReadFull(r, headerBytes)
		require.NoError(t, err)

		header := &pb.BlobHeader{}
		require.NoError(t, proto.Unmarshal(headerBytes, header))

		blobBytes := make([]byte, header.GetDatasize())
		_, err = io.ReadFull(r, blobBytes)
		require.NoError(t, err)

		blob := &pb.Blob{}
		require.NoError(t, proto.Unmarshal(blobBytes, blob))

		blobs = append(blobs, blob)
	}

	return blobs
}

// decodeAll decodes each entity of a PBF stream.
func decodeAll(t *testing.T, encoded []byte) []model.Entity {
	t.Helper()

	r := bytes.NewReader(encoded)

	dec, err := NewDecoder(context.Background(), r)
	require.NoError(t, err)

	defer dec.Close()

	var got []model.Entity

	for {
		batch, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			return got
		}

		require.NoError(t, err)

		got = append(got, batch...)
	}
}
